package utils

import (
	"hash/fnv"
	"net"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func init() {
	var err error
	// 获取基于IP的节点ID
	nodeID := getNodeIDFromIP()
	// 0-1023
	node, err = snowflake.NewNode(nodeID)
	if err != nil {
		panic(err)
	}
}

// getNodeIDFromIP 基于本机IP地址生成节点ID
func getNodeIDFromIP() int64 {
	// 获取本机所有网络接口的IP地址
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		// 如果获取失败，返回默认值
		return 1
	}

	// 查找第一个非本地回环的IPv4地址
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				// 使用FNV哈希算法将IP转换为节点ID
				h := fnv.New32a()
				h.Write(ipnet.IP.To4())
				// 取模确保节点ID在0-1023范围内
				return int64(h.Sum32() % 1024)
			}
		}
	}

	// 如果没有找到合适的IP，返回默认值
	return 1
}

func GenerateID() int64 {
	return node.Generate().Int64()
}
