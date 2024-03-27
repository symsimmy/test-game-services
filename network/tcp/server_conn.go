/**
 * @Author: fuxiao
 * @Email: 576101059@qq.com
 * @Date: 2022/5/11 10:49 上午
 * @Desc: TODO
 */

package tcp

import (
	"bufio"
	"github.com/bytedance/gopkg/util/gopool"
	"github.com/sasha-s/go-deadlock"
	"github.com/symsimmy/due/internal/prom"
	"github.com/symsimmy/due/internal/util"
	"github.com/symsimmy/due/utils/xnet"
	"github.com/symsimmy/due/utils/xtime"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"github.com/symsimmy/due/log"
	"github.com/symsimmy/due/network"
)

const defaultWriteChannelSize = 1024

type serverConn struct {
	rw                deadlock.RWMutex // 锁
	id                int64            // 连接ID
	uid               int64            // 用户ID
	state             int32            // 连接状态
	conn              net.Conn         // TCP源连接
	connMgr           *serverConnMgr   // 连接管理
	chWrite           chan chWrite     // 写入队列
	lastHeartbeatTime int64            // 上次心跳时间
	done              chan struct{}    // 写入完成信号
	reader            *bufio.Reader    // 读取缓冲区
	close             chan struct{}    // 关闭信号
	rBlock            chan struct{}
	rRelease          chan struct{}
	wBlock            chan struct{}
	wRelease          chan struct{}
}

var _ network.Conn = &serverConn{}

// ID 获取连接ID
func (c *serverConn) ID() int64 {
	return c.id
}

// UID 获取用户ID
func (c *serverConn) UID() int64 {
	return atomic.LoadInt64(&c.uid)
}

// Bind 绑定用户ID
func (c *serverConn) Bind(uid int64) {
	atomic.StoreInt64(&c.uid, uid)
}

// Unbind 解绑用户ID
func (c *serverConn) Unbind() {
	atomic.StoreInt64(&c.uid, 0)
}

// Send 发送消息（同步）
func (c *serverConn) Send(msg []byte, msgType ...int) (err error) {
	if err = c.checkState(); err != nil {
		return
	}

	_, err = c.conn.Write(msg)

	return
}

// Push 发送消息（异步）
func (c *serverConn) Push(msg []byte, msgType ...int) (err error) {
	defer func() {
		if r := recover(); r != nil {
		}
	}()

	if err = c.checkState(); err != nil {
		return
	}

	c.chWrite <- chWrite{typ: dataPacket, msg: msg}

	return
}

// State 获取连接状态
func (c *serverConn) State() network.ConnState {
	return network.ConnState(atomic.LoadInt32(&c.state))
}

// Close 关闭连接
func (c *serverConn) Close(isForce ...bool) error {
	if len(isForce) > 0 && isForce[0] {
		return c.forceClose()
	} else {
		return c.graceClose(true)
	}
}

// LocalIP 获取本地IP
func (c *serverConn) LocalIP() (string, error) {
	addr, err := c.LocalAddr()
	if err != nil {
		return "", err
	}

	return xnet.ExtractIP(addr)
}

// LocalAddr 获取本地地址
func (c *serverConn) LocalAddr() (net.Addr, error) {
	if err := c.checkState(); err != nil {
		return nil, err
	}

	return c.conn.LocalAddr(), nil
}

// RemoteIP 获取远端IP
func (c *serverConn) RemoteIP() (string, error) {
	addr, err := c.RemoteAddr()
	if err != nil {
		return "", err
	}

	return xnet.ExtractIP(addr)
}

// RemoteAddr 获取远端地址
func (c *serverConn) RemoteAddr() (net.Addr, error) {
	if err := c.checkState(); err != nil {
		return nil, err
	}

	return c.conn.RemoteAddr(), nil
}

// Block 阻塞conn传来的消息
func (c *serverConn) Block() {
	c.rBlock <- struct{}{}
	c.wBlock <- struct{}{}
}

// Release 释放conn传来的消息
func (c *serverConn) Release() {
	c.rRelease <- struct{}{}
	c.wRelease <- struct{}{}
}

// 检测连接状态
func (c *serverConn) checkState() error {
	switch network.ConnState(atomic.LoadInt32(&c.state)) {
	case network.ConnHanged:
		return network.ErrConnectionHanged
	case network.ConnClosed:
		return network.ErrConnectionClosed
	}

	return nil
}

// 初始化连接
func (c *serverConn) init(conn net.Conn, cm *serverConnMgr) {
	c.id = cm.id
	c.conn = conn
	c.connMgr = cm
	c.chWrite = make(chan chWrite, defaultWriteChannelSize)
	c.done = make(chan struct{})
	c.close = make(chan struct{})
	c.rBlock = make(chan struct{})
	c.rRelease = make(chan struct{})
	c.wBlock = make(chan struct{})
	c.wRelease = make(chan struct{})
	c.lastHeartbeatTime = xtime.Now().Unix()
	atomic.StoreInt64(&c.uid, 0)
	atomic.StoreInt32(&c.state, int32(network.ConnOpened))

	if c.connMgr.server.connectHandler != nil {
		c.connMgr.server.connectHandler(c)
	}
	gopool.Go(c.read)

	gopool.Go(c.write)
}

// 优雅关闭
func (c *serverConn) graceClose(isNeedRecycle bool) (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered in f.r:%+v", r)
			c.conn.Close()
			c.connMgr.recycle(c)
			c.rw.Unlock()
		}
	}()
	c.rw.Lock()

	if err = c.checkState(); err != nil {
		c.rw.Unlock()
		return
	}

	atomic.StoreInt32(&c.state, int32(network.ConnHanged))
	c.chWrite <- chWrite{typ: closeSig}
	c.rw.Unlock()

	<-c.done

	c.rw.Lock()
	atomic.StoreInt32(&c.state, int32(network.ConnClosed))
	close(c.chWrite)
	close(c.done)
	close(c.close)
	c.conn.Close()
	if isNeedRecycle {
		c.connMgr.recycle(c)
	}
	c.rw.Unlock()

	if c.connMgr.server.disconnectHandler != nil {
		c.connMgr.server.disconnectHandler(c)
	}

	return
}

// 强制关闭
func (c *serverConn) forceClose() (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered in f.r:%+v", r)
			c.conn.Close()
			c.connMgr.recycle(c)
			c.rw.Unlock()
		}
	}()
	c.rw.Lock()

	if err = c.checkState(); err != nil {
		c.rw.Unlock()
		return
	}

	atomic.StoreInt32(&c.state, int32(network.ConnClosed))
	close(c.chWrite)
	close(c.done)
	close(c.close)
	c.conn.Close()
	c.connMgr.recycle(c)
	c.rw.Unlock()

	if c.connMgr.server.disconnectHandler != nil {
		c.connMgr.server.disconnectHandler(c)
	}

	return
}

// 读取消息
func (c *serverConn) read() {
	id := c.id
	uid := c.uid
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("cid:%+v,uid%+v,server_conn read task. panic: %v", id, uid, r)
		}
	}()
	c.reader = bufio.NewReader(c.conn)
	for {
		select {
		case <-c.close:
			return
		case <-c.rBlock:
		inner:
			for {
				select {
				case <-c.rRelease:
					break inner
				case <-time.After(3 * time.Second):
					log.Debugf("block cid [%v] uid [%v] client write to server timeout", c.ID(), c.UID())
					_ = c.Close(true)
					break inner
				}
			}
		default:
			msg, err := readMsgFromConn(c.reader, c.connMgr.server.opts.maxMsgLen)
			if err != nil {
				if err == errMsgSizeTooLarge {
					log.Warnf("the msg size too large, has been ignored")
					continue
				}
				if uid > 0 {
					log.Warnf("cid:%+v,uid:%+v,force close.err:%+v", id, uid, err)
				}
				c.forceClose()
				return
			}

			atomic.StoreInt64(&c.lastHeartbeatTime, xtime.Now().Unix())

			switch c.State() {
			case network.ConnHanged:
				continue
			case network.ConnClosed:
				return
			}

			// ignore heartbeat packet
			if len(msg) == 0 {
				continue
			}

			if c.connMgr.server.receiveHandler != nil {
				c.connMgr.server.receiveHandler(c, msg, 0)
			}
		}
	}
}

// 写入消息
func (c *serverConn) write() {
	id := c.id
	uid := c.uid
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("cid:%+v,uid%+v,server_conn write task. panic: %v", id, uid, r)
		}
	}()
	var ticker *time.Ticker
	if c.connMgr.server.opts.enableHeartbeatCheck {
		ticker = time.NewTicker(c.connMgr.server.opts.heartbeatCheckInterval)
		defer ticker.Stop()
	} else {
		ticker = &time.Ticker{C: make(chan time.Time, 1)}
	}

	writeChBacklogTicker := time.NewTicker(1 * time.Second)
	defer writeChBacklogTicker.Stop()

	for {
		select {
		case <-c.wBlock:
		inner:
			for {
				select {
				case <-c.wRelease:
					break inner
				case <-time.After(3 * time.Second):
					log.Debugf("block cid [%v] uid [%v] server write to client timeout", c.ID(), c.UID())
					_ = c.Close(true)
					break inner
				}
			}
		case <-writeChBacklogTicker.C:
			switch c.State() {
			case network.ConnHanged:
				continue
			case network.ConnClosed:
				return
			}
			// 如果积压的数据占满了chWrite，打印log
			if len(c.chWrite) >= defaultWriteChannelSize {
				log.Warnf("cid:%+v,uid:%+v,server connection write channel backlog %+v message", c.id, c.uid, len(c.chWrite))
			}
		case write, ok := <-c.chWrite:
			if !ok {
				return
			}

			if write.typ == closeSig {
				c.done <- struct{}{}
				return
			}

			switch c.State() {
			case network.ConnHanged:
				continue
			case network.ConnClosed:
				return
			}

			buf, err := pack(write.msg)
			if err != nil {
				log.Errorf("packet message error: %v", err)
				continue
			}

			if err = c.doWrite(buf); err != nil {
				if strings.Contains(err.Error(), "connection was aborted") {
					break
				}
			}
		case <-ticker.C:
			deadline := xtime.Now().Add(-2 * c.connMgr.server.opts.heartbeatCheckInterval).Unix()
			if atomic.LoadInt64(&c.lastHeartbeatTime) < deadline {
				log.Infof("connection heartbeat timeout")
				_ = c.Close(true)
				return
			}
		}
	}
}

func (c *serverConn) doWrite(buf []byte) (err error) {
	if atomic.LoadInt32(&c.state) == int32(network.ConnClosed) {
		return
	}
	start := time.Now()
	_, err = c.conn.Write(buf)
	prom.GateServerWriteDurationSummary.WithLabelValues(util.ToString(c.uid), util.ToString(c.id), "").Observe(float64(time.Since(start).Milliseconds()))
	return
}
