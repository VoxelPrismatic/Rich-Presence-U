package ui

import "github.com/godbus/dbus/v5"

// plasmaJob is a Plasma notification progress entry, the same kind file
// transfers use. Nothing calls it yet. The session connection has to stay
// open or Plasma drops the job.
type plasmaJob struct {
	conn *dbus.Conn
	path dbus.ObjectPath
	last int
	msg  string
}

func (p *plasmaJob) ensure(message string) {
	if p.conn == nil {
		conn, err := dbus.ConnectSessionBus()
		if err != nil {
			return
		}
		p.conn = conn
	}
	if p.path != "" {
		return
	}
	var path dbus.ObjectPath
	err := p.conn.Object("org.kde.JobViewServer", "/JobViewServer").Call(
		"org.kde.JobViewServerV2.requestView", 0,
		"rich-presence-u", int32(0), map[string]dbus.Variant{},
	).Store(&path)
	if err != nil || path == "" || path == "/" {
		return
	}
	p.path = path
	p.last = -1
	p.msg = ""
	p.setMessage(message)
}

func (p *plasmaJob) obj() dbus.BusObject {
	if p == nil || p.conn == nil || p.path == "" {
		return nil
	}
	return p.conn.Object("org.kde.JobViewServer", p.path)
}

func (p *plasmaJob) setMessage(message string) {
	if message == p.msg {
		return
	}
	obj := p.obj()
	if obj == nil {
		return
	}
	if obj.Call("org.kde.JobViewV2.setInfoMessage", dbus.FlagNoReplyExpected, message).Err == nil {
		p.msg = message
	}
}

func (p *plasmaJob) setPercent(percent int, message string) {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	p.ensure(message)
	obj := p.obj()
	if obj == nil {
		return
	}
	p.setMessage(message)
	if percent == p.last {
		return
	}
	if obj.Call("org.kde.JobViewV2.setPercent", dbus.FlagNoReplyExpected, uint32(percent)).Err == nil {
		p.last = percent
	}
}

func (p *plasmaJob) close() {
	obj := p.obj()
	if obj != nil {
		_ = obj.Call("org.kde.JobViewV2.setPercent", dbus.FlagNoReplyExpected, uint32(100))
		_ = obj.Call("org.kde.JobViewV2.terminate", dbus.FlagNoReplyExpected, "")
	}
	p.path = ""
	p.last = -1
	p.msg = ""
}
