package xio

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"testing"
	"time"
)

func TestCopy(t *testing.T) {
	t.Run("copy with context cancelation and do not wait", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		unblockWrite := make(chan struct{})
		writeDone := make(chan struct{})

		var writeTotal int

		n, err := Copy(
			ctx,
			WriterFunc(func(b []byte) (int, error) {
				defer close(writeDone)
				cancel()
				<-unblockWrite
				writeTotal += len(b)
				return len(b), nil
			}),
			ReaderFunc(func(b []byte) (int, error) {
				return len(b), nil
			}),
			WaitForLastOp(false),
		)

		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected error to be context canceled but got %v", err)
		}

		if n != 0 {
			t.Fatalf("expected n to be 0 but got %d", n)
		}

		close(unblockWrite)

		<-writeDone

		if writeTotal != 32768 {
			t.Fatalf("expected write total to be 32768 but got %d", writeTotal)
		}
	})

	t.Run("Copy with wait for last write", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())

		unblockWrite := make(chan struct{})

		time.AfterFunc(20*time.Millisecond, func() { close(unblockWrite) })

		n, err := Copy(
			ctx,
			WriterFunc(func(b []byte) (int, error) {
				cancel()
				<-unblockWrite
				return len(b), nil
			}),
			ReaderFunc(func(b []byte) (int, error) {
				return len(b), nil
			}),
			WaitForLastOp(true),
		)

		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected error to be context canceled but got %v", err)
		}

		if n != 32768 {
			t.Fatalf("expected n to be 32768 but got %d", n)
		}
	})

	t.Run("canceled write error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		writeErr := errors.New("writer broke!")

		n, err := Copy(
			ctx,
			WriterFunc(func(b []byte) (int, error) { return 42, writeErr }),
			ReaderFunc(func(b []byte) (int, error) {
				cancel()
				time.Sleep(20 * time.Millisecond) // we give the time for the context.Done to have fired
				return len(b), nil
			}),
		)

		if err != writeErr {
			t.Fatalf("expected write err to be %#q but got %#q", writeErr, err)
		}
		if n != 42 {
			t.Fatalf("expected n to be 42 but got %d", n)
		}
	})

	t.Run("canceled read error", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		readErr := errors.New("reader broke!")

		n, err := Copy(
			ctx,
			WriterFunc(func(b []byte) (int, error) { return 42, nil }),
			ReaderFunc(func(b []byte) (int, error) {
				cancel()
				time.Sleep(20 * time.Millisecond) // we give the time for the context.Done to have fired
				return 0, readErr
			}),
		)

		if err != readErr {
			t.Fatalf("expected read err to be %#q but got %#q", readErr, err)
		}
		if n != 0 {
			t.Fatalf("expected n to be 0 but got %d", n)
		}
	})

	t.Run("calling with canceled context is a noop", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// will panic if Copy tries to read from src
		n, err := Copy(ctx, nil, nil)

		if err != context.Canceled {
			t.Fatalf("expected err to be %#q but got %#q", context.Canceled, err)
		}
		if n != 0 {
			t.Fatalf("expected n to be 0 but got %d", n)
		}
	})

	t.Run("read error will write what it can then return error", func(t *testing.T) {
		readErr := errors.New("reader encoutered invalid state!")

		n, err := Copy(
			context.Background(),
			WriterFunc(func(b []byte) (int, error) { return len(b), nil }),
			ReaderFunc(func(b []byte) (int, error) { return 42, readErr }),
		)
		if err != readErr {
			t.Fatalf("expected err to be %#q but got %#q", readErr, err)
		}
		if n != 42 {
			t.Fatalf("expected n to be 42 but got %d", n)
		}
	})

	t.Run("invalid write err", func(t *testing.T) {
		n, err := Copy(
			context.Background(),
			WriterFunc(func(b []byte) (int, error) { return 100, nil }),
			ReaderFunc(func(b []byte) (int, error) { return 42, nil }),
		)
		if err != errInvalidWrite {
			t.Fatalf("expected err to be %#q but got %#q", errInvalidWrite, err)
		}
		if n != 0 {
			t.Fatalf("expected n to be 0 but got %d", n)
		}
	})

	t.Run("can set buffer size", func(t *testing.T) {
		var called bool

		n, err := Copy(
			context.Background(),
			WriterFunc(func(b []byte) (int, error) {
				return len(b), nil
			}),
			ReaderFunc(func(b []byte) (int, error) {
				called = true
				return len(b), io.EOF
			}),
			BufferSize(16),
		)

		if !called {
			t.Fatal("Copy read func was not called")
		}
		if err != nil {
			t.Fatalf("expected err to be nil but got %#q", err)
		}
		if n != 16 {
			t.Fatalf("expected n to be buffersize %d but got %d", 16, n)
		}
	})
}

func TestCopyN(t *testing.T) {
	t.Run("only copies N bytes", func(t *testing.T) {
		var bytesRead []int
		n, err := CopyN(
			context.Background(),
			WriterFunc(func(b []byte) (int, error) { return len(b), nil }),
			ReaderFunc(func(b []byte) (int, error) {
				bytesRead = append(bytesRead, len(b))
				return len(b), nil
			}),
			100,
			// we want multiple reads with one that overflows so we would have technically read 128 bytes
			// but we will still get a n of 100
			BufferSize(64),
		)
		if err != nil {
			t.Fatalf("expected err to be nil but got %#q", err)
		}
		if n != 100 {
			t.Fatalf("expected to total written to be 100 but got %d", n)
		}

		expectedBytesRead := []int{64, 36}
		if !reflect.DeepEqual(bytesRead, expectedBytesRead) {
			t.Fatalf("expected bytes to be read like so %v but got %v", expectedBytesRead, bytesRead)
		}
	})

	t.Run("read less data than N", func(t *testing.T) {
		n, err := CopyN(
			context.Background(),
			WriterFunc(func(b []byte) (int, error) { return len(b), nil }),
			ReaderFunc(func(b []byte) (int, error) { return 50, io.EOF }),
			100,
		)

		if err != io.EOF {
			t.Fatalf("expected err to be %#q but got %#q", io.EOF, err)
		}
		if n != 50 {
			t.Fatalf("expected n to be 50 but got %d", n)
		}
	})

	t.Run("N of zero should not trigger any reads", func(t *testing.T) {
		// Will panic if src Read is called
		n, err := CopyN(context.Background(), nil, nil, 0)
		if err != nil {
			t.Fatalf("expected err to be nil but got %#q", err)
		}
		if n != 0 {
			t.Fatalf("expected n to be 0 but got %d", n)
		}
	})
}

func TestCopyBuffer(t *testing.T) {
	buffer := make([]byte, 15)

	n, err := CopyBuffer(
		context.Background(),
		WriterFunc(func(b []byte) (int, error) {
			return len(b), nil
		}),
		ReaderFunc(func(b []byte) (int, error) {
			for i, v := range []byte("hello world") {
				b[i] = v
			}
			return 11, io.EOF
		}),
		buffer,
	)
	if err != nil {
		t.Fatalf("expected no error but got %v", err)
	}

	if n != 11 {
		t.Fatalf("expected 11 bytes to be read but got: %d", n)
	}

	if substring := string(buffer[:11]); substring != "hello world" {
		t.Fatalf("expected hello world but got: %s", substring)
	}
}

func TestReadAll(t *testing.T) {
	expected := []byte(`Hello world`)

	actual, err := ReadAll(context.Background(), bytes.NewReader(expected))
	if err != nil {
		t.Fatalf("expectd err to be nil but got %#q", err)
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("expect content to be %q but got %q", expected, actual)
	}
}

type ReaderFunc func([]byte) (int, error)

func (fn ReaderFunc) Read(data []byte) (int, error) { return fn(data) }

type WriterFunc func([]byte) (int, error)

func (fn WriterFunc) Write(data []byte) (int, error) { return fn(data) }
