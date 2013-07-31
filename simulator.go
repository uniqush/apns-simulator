package main

type SimulatorFactory interface {
	MakeSimulator() (sim APNSSimulator, err error)
}

type APNSSimulator interface {
	Reply(notif *APNSNotificaton) (res *APNSResponse, err error)
}

type APNSDefinedBehaviorSimulator struct {
	Statuses []uint8
	idx      int
}

type StatusSimulatorFactory struct {
	Statuses []uint8
}

func NewStatusSimulatorFactory(status ...uint8) SimulatorFactory {
	ret := new(StatusSimulatorFactory)
	if len(status) > 0 {
		ret.Statuses = make([]uint8, len(status))
		for i, s := range status {
			ret.Statuses[i] = s
		}
	}
	return ret
}

func (self *StatusSimulatorFactory) MakeSimulator() (sim APNSSimulator, err error) {
	sim = NewAPNSSimulatorWithStatuses(self.Statuses...)
	return
}

type NormalSimulatorFactory struct {
	MaxPayloadLen  int
	DeviceTokenLen int
}

func NewNormalSimulatorFactory(MaxPayloadLen, DeviceTokenLen int) SimulatorFactory {
	ret := &NormalSimulatorFactory{MaxPayloadLen, DeviceTokenLen}
	return ret
}

func (self *NormalSimulatorFactory) MakeSimulator() (sim APNSSimulator, err error) {
	sim = &APNSNormalSimulator{self.MaxPayloadLen, self.DeviceTokenLen}
	return
}

func NewAPNSSimulatorWithStatuses(status ...uint8) *APNSDefinedBehaviorSimulator {
	ret := new(APNSDefinedBehaviorSimulator)
	if len(status) > 0 {
		ret.Statuses = make([]uint8, len(status))
		for i, s := range status {
			ret.Statuses[i] = s
		}
	}
	return ret
}

func (self *APNSDefinedBehaviorSimulator) Reply(notif *APNSNotificaton) (res *APNSResponse, err error) {
	res = new(APNSResponse)
	res.id = notif.id
	if len(self.Statuses) == 0 {
		return
	}
	if self.idx >= len(self.Statuses) || self.idx < 0 {
		self.idx = 0
	}
	res.status = self.Statuses[self.idx]
	self.idx++
	return
}

type APNSNormalSimulator struct {
	MaxPayloadLen  int
	DeviceTokenLen int
}

func (self *APNSNormalSimulator) Reply(notif *APNSNotificaton) (res *APNSResponse, err error) {
	res = new(APNSResponse)
	res.id = notif.id
	if self.MaxPayloadLen <= 0 {
		self.MaxPayloadLen = 256
	}
	if self.DeviceTokenLen <= 0 {
		self.DeviceTokenLen = 32
	}
	if len(notif.payload) > self.MaxPayloadLen {
		res.status = 7
	} else if len(notif.devToken) != self.DeviceTokenLen {
		res.status = 5
	}
	// TODO detect more errors
	return
}
