package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/sourcegraph/jsonrpc2"
)

const (
	TDSKNone        TextDocumentSyncKind = 0
	TDSKFull                             = 1
	TDSKIncremental                      = 2
)

type Server struct {
	files map[string]*File
}

type File struct {
	LanguageID string
	Text       string
}

func NewServer() *Server {
	return &Server{
		files: make(map[string]*File),
	}
}

func (s *Server) handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(ctx, conn, req)
	case "initialized":
		return
	case "shutdown":
		return s.handleShutdown(ctx, conn, req)
	case "textDocument/didOpen":
		return s.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didChange":
		return s.handleTextDocumentDidChange(ctx, conn, req)
	case "textDocument/didSave":
		return s.handleTextDocumentDidSave(ctx, conn, req)
	case "textDocument/didClose":
		return s.handleTextDocumentDidClose(ctx, conn, req)
	case "textDocument/completion":
		return s.handleTextDocumentCompletion(ctx, conn, req)
		// case "textDocument/formatting":
		// 	return h.handleTextDocumentFormatting(ctx, conn, req)
		// case "textDocument/documentSymbol":
		// 	return h.handleTextDocumentSymbol(ctx, conn, req)
	}

	return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeMethodNotFound, Message: fmt.Sprintf("method not supported: %s", req.Method)}
}

func (s *Server) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	return InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: TDSKFull,
			HoverProvider:    false,
			CompletionProvider: &CompletionOptions{
				TriggerCharacters: []string{"."},
			},
			DefinitionProvider:              false,
			DocumentFormattingProvider:      false,
			DocumentRangeFormattingProvider: false,
		},
	}, nil
}

func (s *Server) handleShutdown(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}
	return nil, nil
}

func (s *Server) handleTextDocumentDidOpen(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	log.Println("handle textdocument/didOpen")
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	s.openFile(params.TextDocument.URI, params.TextDocument.LanguageID)
	if err := s.updateFile(params.TextDocument.URI, params.TextDocument.Text); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Server) handleTextDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	log.Println("handle textdocument/didChange")
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	if err := s.updateFile(params.TextDocument.URI, params.ContentChanges[0].Text); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Server) handleTextDocumentDidSave(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	log.Println("handle textdocument/didSave")
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params DidSaveTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	if params.Text != "" {
		err = s.updateFile(params.TextDocument.URI, params.Text)
	} else {
		err = s.saveFile(params.TextDocument.URI)
	}
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Server) handleTextDocumentDidClose(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	log.Println("handle textdocument/didClose")
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params DidCloseTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	s.closeFile(params.TextDocument.URI)
	return nil, nil
}

func (s *Server) handleTextDocumentCompletion(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (result interface{}, err error) {
	log.Println("handle textDocument/completion")
	if req.Params == nil {
		return nil, &jsonrpc2.Error{Code: jsonrpc2.CodeInvalidParams}
	}

	var params CompletionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return nil, err
	}

	completer := &Completer{}
	completionItems, err := completer.complete(params)
	if err != nil {
		return nil, err
	}
	result = completionItems
	return
}

func (s *Server) openFile(uri string, languageID string) error {
	f := &File{
		Text:       "",
		LanguageID: languageID,
	}
	s.files[uri] = f
	return nil
}

func (s *Server) closeFile(uri string) error {
	delete(s.files, uri)
	return nil
}

func (s *Server) updateFile(uri string, text string) error {
	f, ok := s.files[uri]
	if !ok {
		return fmt.Errorf("document not found: %v", uri)
	}
	f.Text = text
	return nil
}

func (s *Server) saveFile(uri string) error {
	return nil
}
