package catalog

import "testing"

// TestReimbursementComponentsUsesIndependentIdentity 防止报销与其他业务共享运行许可组件身份。
// 参数：t为测试上下文。
// 返回值：无。
func TestReimbursementComponentsUsesIndependentIdentity(t *testing.T) {
	components, err := Components("reimbursement")
	if err != nil {
		t.Fatal(err)
	}
	if len(components) != 2 {
		t.Fatalf("expected API and worker, got %d", len(components))
	}
	for _, id := range []string{"reimbursement-api", "reimbursement-worker"} {
		c, err := Lookup("reimbursement", id)
		if err != nil || c.RegistrationProtocol != RegistrationProtocolV2 {
			t.Fatalf("%s: %+v %v", id, c, err)
		}
	}
	if _, err = Lookup("contract_management", "reimbursement-api"); err == nil {
		t.Fatal("cross application lookup accepted")
	}
}
