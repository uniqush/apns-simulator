package main

type APNSSimulator interface {
	Reply(notif *APNSNotificaton) (res *APNSResponse, err error)
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
