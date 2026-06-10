package controllers

import "testing"

func TestExtractJushuitanOrderID(t *testing.T) {
	resp := `{"msg":"执行成功","code":0,"data":{"datas":[{"msg":"成功","issuccess":true,"so_id":"Y2026061081412847","o_id":5896709,"split_id":null}],"requestId":null}}`

	got, err := extractJushuitanOrderID(resp, "Y2026061081412847")
	if err != nil {
		t.Fatalf("extractJushuitanOrderID returned error: %v", err)
	}
	if got != "5896709" {
		t.Fatalf("expected 5896709, got %q", got)
	}
}

func TestExtractJushuitanOrderIDRequiresSuccessfulOID(t *testing.T) {
	resp := `{"msg":"执行成功","code":0,"data":{"datas":[{"msg":"失败","issuccess":false,"so_id":"Y001","o_id":null}],"requestId":null}}`

	if _, err := extractJushuitanOrderID(resp, "Y001"); err == nil {
		t.Fatalf("expected error for failed upload item")
	}
}
