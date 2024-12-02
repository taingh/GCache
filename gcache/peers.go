package gcache

import pb "gcache/gcachepb"

// PeerPicker 是必须实现的接口，用于定位拥有特定键的对等节点。
// PickPeer 根据给定的键选择一个对等节点。
//   - key: 要查找的键
//   - 返回值:
//     - peer: 实现了 PeerGetter 接口的对等节点
//     - ok: 是否成功找到对等节点
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// PeerGetter 是必须由对等节点实现的接口。
// Get 方法从对等节点获取数据。
//   - in: 请求对象，包含请求的键
//   - out: 响应对象，包含返回的数据
//   - 返回值:
//     - error: 操作过程中发生的错误
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}