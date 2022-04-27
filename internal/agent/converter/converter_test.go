package converter

import (
	"reflect"
	"testing"

	"github.com/scrolllockdev/test-devops/internal/agent/storage"
)

func TestGaugeToJSON(t *testing.T) {
	type args struct {
		name  string
		value storage.Gauge
		key   string
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test 1",
			args: struct {
				name  string
				value storage.Gauge
				key   string
			}{name: "MCacheSys", value: 37536, key: "as"},
			want:    []byte{123, 34, 105, 100, 34, 58, 34, 77, 67, 97, 99, 104, 101, 83, 121, 115, 34, 44, 34, 116, 121, 112, 101, 34, 58, 34, 103, 97, 117, 103, 101, 34, 44, 34, 118, 97, 108, 117, 101, 34, 58, 51, 55, 53, 51, 54, 44, 34, 104, 97, 115, 104, 34, 58, 34, 56, 50, 101, 98, 99, 97, 101, 48, 53, 56, 52, 98, 51, 54, 56, 52, 102, 100, 57, 54, 56, 49, 51, 48, 54, 102, 51, 100, 98, 56, 53, 56, 53, 54, 55, 101, 52, 48, 54, 101, 98, 51, 52, 101, 55, 48, 100, 97, 50, 53, 98, 51, 101, 99, 102, 98, 55, 50, 100, 56, 99, 99, 98, 55, 34, 125},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GaugeToJSON(tt.args.name, tt.args.value, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("GaugeToJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GaugeToJSON() got = %v, want %v", got, tt.want)
			}
		})
	}
}
