package data

import "github.com/mediocregopher/radix/v3"

type cacheCluster struct {
	cluster *radix.Cluster
}

func NewCacheClusterClient(addresses []string, password string) (CacheClient, error) {
	connFunc := func(network string, addr string) (radix.Conn, error) {
		return radix.Dial(
			network,
			addr,
			radix.DialAuthPass(password),
		)
	}
	poolFunc := func(network, addr string) (radix.Client, error) {
		return radix.NewPool(network, addr, 10, radix.PoolConnFunc(connFunc))
	}
	cluster, err := radix.NewCluster(addresses, radix.ClusterPoolFunc(poolFunc))
	if err != nil {
		return nil, err
	}

	return &cacheCluster{
		cluster: cluster,
	}, nil
}

func (r cacheCluster) Del(keys ...string) error {
	return r.cluster.Do(radix.Cmd(nil, "DEL", keys...))
}

func (r cacheCluster) HSet(key, field string, value string) error {
	return r.cluster.Do(radix.Cmd(nil, "HSET", key, field, value))
}

func (r cacheCluster) HGet(key, field string) (*string, error) {
	var result string
	value := radix.MaybeNil{
		Rcv: &result,
	}
	if err := r.cluster.Do(radix.Cmd(&value, "HGET", key, field)); err != nil {
		return nil, err
	}
	if value.Nil {
		return nil, nil
	}
	return &result, nil
}

func (r cacheCluster) HDel(key string, fields ...string) error {
	args := make([]string, 0)
	args = append(args, key)
	args = append(args, fields...)

	return r.cluster.Do(radix.Cmd(nil, "HDEL", args...))
}

func (r cacheCluster) HGetAll(key string) (map[string]string, error) {
	var result map[string]string
	value := radix.MaybeNil{
		Rcv: &result,
	}
	err := r.cluster.Do(radix.Cmd(&value, "HGETALL", key))
	if err != nil {
		return nil, err
	}
	if value.Nil {
		return nil, nil
	}
	return result, nil
}

func (r cacheCluster) HMSet(key string, values map[string]string) error {
	args := make([]string, 0)
	args = append(args, key)
	for k, v := range values {
		args = append(args, k)
		args = append(args, v)

		if len(args) > multiSetStepLimit {
			if err := r.cluster.Do(radix.Cmd(nil, "HMSET", args...)); err != nil {
				return err
			}
			args = make([]string, 0)
			args = append(args, key)
		}
	}
	if len(args) == 1 {
		return nil
	}
	return r.cluster.Do(radix.Cmd(nil, "HMSET", args...))
}

func (r cacheCluster) Pipeline(commands []radix.CmdAction) error {
	if len(commands) == 0 {
		return nil
	}
	return r.cluster.Do(radix.Pipeline(commands...))
}

func (r cacheCluster) Get(key string) (*string, error) {
	var result string
	value := radix.MaybeNil{
		Rcv: &result,
	}
	if err := r.cluster.Do(radix.Cmd(&value, "GET", key)); err != nil {
		return nil, err
	}
	if value.Nil {
		return nil, nil
	}
	return &result, nil
}

func (r cacheCluster) Set(key string, value string) error {
	return r.cluster.Do(radix.Cmd(nil, "SET", key, value))
}

func (r cacheCluster) Do(cmd radix.CmdAction) error {
	return r.cluster.Do(cmd)
}

var _ CacheClient = &cacheCluster{}
