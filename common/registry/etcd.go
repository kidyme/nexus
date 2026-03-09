package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultKeyPrefix       = "/nexus/meta"
	defaultLeaseTTL  int64 = 10
)

// EtcdRegistry implements service registration on top of etcd.
type EtcdRegistry struct {
	client *clientv3.Client
	config Config
	node   *Node
	lease  clientv3.LeaseID
}

// NewEtcdRegistry creates a registry backed by etcd.
func NewEtcdRegistry(cfg Config) (*EtcdRegistry, error) {
	if len(cfg.Endpoints) == 0 {
		return nil, ErrEndpointsRequired
	}
	if cfg.DialTimeout <= 0 {
		cfg.DialTimeout = 5 * time.Second
	}
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 3 * time.Second
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = defaultKeyPrefix
	}
	if cfg.LeaseTTL <= 0 {
		cfg.LeaseTTL = defaultLeaseTTL
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   cfg.Endpoints,
		Username:    cfg.Username,
		Password:    cfg.Password,
		DialTimeout: cfg.DialTimeout,
	})
	if err != nil {
		return nil, err
	}

	return &EtcdRegistry{
		client: client,
		config: cfg,
	}, nil
}

// Register registers the current node with a lease.
func (r *EtcdRegistry) Register(ctx context.Context, node Node) error {
	if err := validateNode(node); err != nil {
		return err
	}

	node.HeartbeatAt = time.Now().UTC()
	leaseCtx, cancel := r.withRequestTimeout(ctx)
	defer cancel()

	leaseResp, err := r.client.Grant(leaseCtx, r.config.LeaseTTL)
	if err != nil {
		return err
	}

	if err := r.putNode(ctx, node, leaseResp.ID); err != nil {
		return err
	}

	r.node = &node
	r.lease = leaseResp.ID
	return nil
}

// Heartbeat refreshes the node lease and heartbeat timestamp.
func (r *EtcdRegistry) Heartbeat(ctx context.Context) error {
	if r.node == nil || r.lease == 0 {
		return ErrNodeNotFound
	}

	keepAliveCtx, cancel := r.withRequestTimeout(ctx)
	defer cancel()
	if _, err := r.client.KeepAliveOnce(keepAliveCtx, r.lease); err != nil {
		return err
	}

	node := *r.node
	node.HeartbeatAt = time.Now().UTC()
	if err := r.putNode(ctx, node, r.lease); err != nil {
		return err
	}

	r.node = &node
	return nil
}

// Discover lists alive nodes under a service prefix.
func (r *EtcdRegistry) Discover(ctx context.Context, serviceName string) ([]Node, error) {
	getCtx, cancel := r.withRequestTimeout(ctx)
	defer cancel()

	resp, err := r.client.Get(getCtx, servicePrefix(r.config.KeyPrefix, serviceName), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	nodes := make([]Node, 0, len(resp.Kvs))
	for _, kv := range resp.Kvs {
		var node Node
		if err := json.Unmarshal(kv.Value, &node); err != nil {
			return nil, err
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

// Probe checks whether a single node is still alive in etcd.
func (r *EtcdRegistry) Probe(ctx context.Context, serviceName, nodeID string) (*Node, error) {
	getCtx, cancel := r.withRequestTimeout(ctx)
	defer cancel()

	resp, err := r.client.Get(getCtx, nodeKey(r.config.KeyPrefix, serviceName, nodeID))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return nil, ErrNodeNotFound
	}

	var node Node
	if err := json.Unmarshal(resp.Kvs[0].Value, &node); err != nil {
		return nil, err
	}
	return &node, nil
}

// Deregister deletes the current node and revokes the lease.
func (r *EtcdRegistry) Deregister(ctx context.Context) error {
	if r.node == nil || r.lease == 0 {
		return nil
	}

	deleteCtx, cancel := r.withRequestTimeout(ctx)
	defer cancel()
	if _, err := r.client.Delete(deleteCtx, nodeKey(r.config.KeyPrefix, r.node.ServiceName, r.node.NodeID)); err != nil {
		return err
	}

	revokeCtx, revokeCancel := r.withRequestTimeout(ctx)
	defer revokeCancel()
	if _, err := r.client.Revoke(revokeCtx, r.lease); err != nil {
		return err
	}

	r.node = nil
	r.lease = 0
	return nil
}

// Close closes the underlying etcd client.
func (r *EtcdRegistry) Close() error {
	return r.client.Close()
}

func (r *EtcdRegistry) putNode(ctx context.Context, node Node, lease clientv3.LeaseID) error {
	payload, err := json.Marshal(node)
	if err != nil {
		return err
	}

	putCtx, cancel := r.withRequestTimeout(ctx)
	defer cancel()
	_, err = r.client.Put(putCtx, nodeKey(r.config.KeyPrefix, node.ServiceName, node.NodeID), string(payload), clientv3.WithLease(lease))
	return err
}

func (r *EtcdRegistry) withRequestTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, r.config.RequestTimeout)
}

func validateNode(node Node) error {
	switch {
	case node.ServiceName == "":
		return fmt.Errorf("registry: service name is required")
	case node.NodeID == "":
		return fmt.Errorf("registry: node id is required")
	case node.Endpoint == "":
		return fmt.Errorf("registry: endpoint is required")
	}
	return nil
}

func servicePrefix(keyPrefix, serviceName string) string {
	return path.Join(cleanKeyPrefix(keyPrefix), "nodes", serviceName) + "/"
}

func nodeKey(keyPrefix, serviceName, nodeID string) string {
	return path.Join(cleanKeyPrefix(keyPrefix), "nodes", serviceName, nodeID)
}

func cleanKeyPrefix(keyPrefix string) string {
	keyPrefix = strings.TrimSpace(keyPrefix)
	if keyPrefix == "" {
		return defaultKeyPrefix
	}
	if strings.HasPrefix(keyPrefix, "/") {
		return path.Clean(keyPrefix)
	}
	return "/" + path.Clean(keyPrefix)
}
