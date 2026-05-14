package auth

import "context"

type VINWhitelist struct {
	set map[string]struct{}
}

func NewVINWhitelist(vins []string) *VINWhitelist {
	w := &VINWhitelist{
		set: make(map[string]struct{}, len(vins)),
	}
	for _, v := range vins {
		w.set[v] = struct{}{}
	}
	return w
}

func (w *VINWhitelist) Add(vin string) {
	w.set[vin] = struct{}{}
}

func (w *VINWhitelist) Remove(vin string) {
	delete(w.set, vin)
}

func (w *VINWhitelist) Authenticate(ctx context.Context, vin string) (bool, error) {
	_, ok := w.set[vin]
	return ok, nil
}

type AllowAll struct{}

func (a *AllowAll) Authenticate(ctx context.Context, vin string) (bool, error) {
	return true, nil
}
