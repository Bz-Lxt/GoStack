package control

import "testing"

func TestImpairDropAndStep(t *testing.T) {
	im := NewImpair()
	im.Configure(1, 0, 0, 0)
	if !im.Drop() {
		t.Fatal("loss=1 must drop")
	}
	st := NewStep(20)
	st.Set(false, 5)
	st.Wait()
	if st.Snapshot()["enabled"] != false {
		t.Fatal(st.Snapshot())
	}
}
