// Package redisx 提供极简 Redis 访问能力。
package redisx

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

// Options 定义 Redis 连接选项。
type Options struct {
	Addr         string
	Password     string
	DB           int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Client 是极简 Redis 客户端。
type Client struct {
	addr         string
	password     string
	db           int
	dialTimeout  time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration
}

// NewClient 创建 Redis 客户端。
func NewClient(options Options) *Client {
	if options.Addr == "" {
		options.Addr = "127.0.0.1:6379"
	}
	if options.DialTimeout <= 0 {
		options.DialTimeout = 3 * time.Second
	}
	if options.ReadTimeout <= 0 {
		options.ReadTimeout = 3 * time.Second
	}
	if options.WriteTimeout <= 0 {
		options.WriteTimeout = 3 * time.Second
	}
	return &Client{
		addr:         options.Addr,
		password:     options.Password,
		db:           options.DB,
		dialTimeout:  options.DialTimeout,
		readTimeout:  options.ReadTimeout,
		writeTimeout: options.WriteTimeout,
	}
}

// Ping 检查 Redis 是否可连通。
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.exec(ctx, command{name: "PING"})
	return err
}

// Set 写入单个字符串值。
func (c *Client) Set(ctx context.Context, key, value string) error {
	_, err := c.exec(ctx, command{name: "SET", args: []string{key, value}})
	return err
}

// Get 读取单个字符串值。
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	replies, err := c.exec(ctx, command{name: "GET", args: []string{key}})
	if err != nil {
		return "", err
	}
	if len(replies) == 0 {
		return "", nil
	}
	return replies[0], nil
}

// SetMany 批量写入多个字符串值。
func (c *Client) SetMany(ctx context.Context, values map[string]string) error {
	if len(values) == 0 {
		return nil
	}
	commands := make([]command, 0, len(values))
	for key, value := range values {
		commands = append(commands, command{name: "SET", args: []string{key, value}})
	}
	_, err := c.exec(ctx, commands...)
	return err
}

type command struct {
	name string
	args []string
}

func (c *Client) exec(ctx context.Context, commands ...command) ([]string, error) {
	conn, err := c.dial(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	setup := make([]command, 0, 2+len(commands))
	if c.password != "" {
		setup = append(setup, command{name: "AUTH", args: []string{c.password}})
	}
	if c.db > 0 {
		setup = append(setup, command{name: "SELECT", args: []string{strconv.Itoa(c.db)}})
	}
	setup = append(setup, commands...)

	if err := setDeadline(conn, c.writeTimeout, ctx); err != nil {
		return nil, err
	}
	for _, cmd := range setup {
		if _, err := writer.Write(encodeCommand(cmd)); err != nil {
			return nil, err
		}
	}
	if err := writer.Flush(); err != nil {
		return nil, err
	}

	if err := setDeadline(conn, c.readTimeout, ctx); err != nil {
		return nil, err
	}
	replies := make([]string, 0, len(setup))
	for range setup {
		reply, err := readReply(reader)
		if err != nil {
			return nil, err
		}
		replies = append(replies, reply)
	}

	offset := len(replies) - len(commands)
	if offset < 0 {
		offset = 0
	}
	return replies[offset:], nil
}

func (c *Client) dial(ctx context.Context) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: c.dialTimeout}
	return dialer.DialContext(ctx, "tcp", c.addr)
}

func encodeCommand(cmd command) []byte {
	parts := make([]string, 0, len(cmd.args)+1)
	parts = append(parts, cmd.name)
	parts = append(parts, cmd.args...)

	var builder strings.Builder
	builder.WriteString("*")
	builder.WriteString(strconv.Itoa(len(parts)))
	builder.WriteString("\r\n")
	for _, part := range parts {
		builder.WriteString("$")
		builder.WriteString(strconv.Itoa(len(part)))
		builder.WriteString("\r\n")
		builder.WriteString(part)
		builder.WriteString("\r\n")
	}
	return []byte(builder.String())
}

func readReply(reader *bufio.Reader) (string, error) {
	prefix, err := reader.ReadByte()
	if err != nil {
		return "", err
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(line, "\n")
	line = strings.TrimSuffix(line, "\r")

	switch prefix {
	case '+', ':':
		return line, nil
	case '-':
		return "", errors.New(line)
	case '$':
		size, err := strconv.Atoi(line)
		if err != nil {
			return "", err
		}
		if size < 0 {
			return "", nil
		}
		data := make([]byte, size+2)
		if _, err := io.ReadFull(reader, data); err != nil {
			return "", err
		}
		return string(data[:size]), nil
	default:
		return "", fmt.Errorf("redisx: unsupported reply prefix %q", prefix)
	}
}

func setDeadline(conn net.Conn, timeout time.Duration, ctx context.Context) error {
	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		deadline = ctxDeadline
	}
	return conn.SetDeadline(deadline)
}
