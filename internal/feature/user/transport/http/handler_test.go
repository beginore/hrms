package http_test

// The crucial point here is to write all test in their own package with '_test' postfix.
// This is needed to avoid cycle imports.
// For example with such configuration you can freely call mocks from http/mock directory (mock package)
// and use real logic form http/handler.go (http package)
