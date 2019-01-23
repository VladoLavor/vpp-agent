// Code generated by adapter-generator. DO NOT EDIT.

package adapter

import (
	"github.com/gogo/protobuf/proto"
	. "github.com/ligato/vpp-agent/plugins/kvscheduler/api"
	"github.com/ligato/vpp-agent/pkg/idxvpp2"
	"github.com/ligato/vpp-agent/api/models/vpp/l2"
)

////////// type-safe key-value pair with metadata //////////

type BridgeDomainKVWithMetadata struct {
	Key      string
	Value    *vpp_l2.BridgeDomain
	Metadata *idxvpp2.OnlyIndex
	Origin   ValueOrigin
}

////////// type-safe Descriptor structure //////////

type BridgeDomainDescriptor struct {
	Name               string
	KeySelector        KeySelector
	ValueTypeName      string
	KeyLabel           func(key string) string
	ValueComparator    func(key string, oldValue, newValue *vpp_l2.BridgeDomain) bool
	NBKeyPrefix        string
	WithMetadata       bool
	MetadataMapFactory MetadataMapFactory
	Add                func(key string, value *vpp_l2.BridgeDomain) (metadata *idxvpp2.OnlyIndex, err error)
	Delete             func(key string, value *vpp_l2.BridgeDomain, metadata *idxvpp2.OnlyIndex) error
	Modify             func(key string, oldValue, newValue *vpp_l2.BridgeDomain, oldMetadata *idxvpp2.OnlyIndex) (newMetadata *idxvpp2.OnlyIndex, err error)
	ModifyWithRecreate func(key string, oldValue, newValue *vpp_l2.BridgeDomain, metadata *idxvpp2.OnlyIndex) bool
	Update             func(key string, value *vpp_l2.BridgeDomain, metadata *idxvpp2.OnlyIndex) error
	IsRetriableFailure func(err error) bool
	Dependencies       func(key string, value *vpp_l2.BridgeDomain) []Dependency
	DerivedValues      func(key string, value *vpp_l2.BridgeDomain) []KeyValuePair
	Dump               func(correlate []BridgeDomainKVWithMetadata) ([]BridgeDomainKVWithMetadata, error)
	DumpDependencies   []string /* descriptor name */
}

////////// Descriptor adapter //////////

type BridgeDomainDescriptorAdapter struct {
	descriptor *BridgeDomainDescriptor
}

func NewBridgeDomainDescriptor(typedDescriptor *BridgeDomainDescriptor) *KVDescriptor {
	adapter := &BridgeDomainDescriptorAdapter{descriptor: typedDescriptor}
	descriptor := &KVDescriptor{
		Name:               typedDescriptor.Name,
		KeySelector:        typedDescriptor.KeySelector,
		ValueTypeName:      typedDescriptor.ValueTypeName,
		KeyLabel:           typedDescriptor.KeyLabel,
		NBKeyPrefix:        typedDescriptor.NBKeyPrefix,
		WithMetadata:       typedDescriptor.WithMetadata,
		MetadataMapFactory: typedDescriptor.MetadataMapFactory,
		IsRetriableFailure: typedDescriptor.IsRetriableFailure,
		DumpDependencies:   typedDescriptor.DumpDependencies,
	}
	if typedDescriptor.ValueComparator != nil {
		descriptor.ValueComparator = adapter.ValueComparator
	}
	if typedDescriptor.Add != nil {
		descriptor.Add = adapter.Add
	}
	if typedDescriptor.Delete != nil {
		descriptor.Delete = adapter.Delete
	}
	if typedDescriptor.Modify != nil {
		descriptor.Modify = adapter.Modify
	}
	if typedDescriptor.ModifyWithRecreate != nil {
		descriptor.ModifyWithRecreate = adapter.ModifyWithRecreate
	}
	if typedDescriptor.Update != nil {
		descriptor.Update = adapter.Update
	}
	if typedDescriptor.Dependencies != nil {
		descriptor.Dependencies = adapter.Dependencies
	}
	if typedDescriptor.DerivedValues != nil {
		descriptor.DerivedValues = adapter.DerivedValues
	}
	if typedDescriptor.Dump != nil {
		descriptor.Dump = adapter.Dump
	}
	return descriptor
}

func (da *BridgeDomainDescriptorAdapter) ValueComparator(key string, oldValue, newValue proto.Message) bool {
	typedOldValue, err1 := castBridgeDomainValue(key, oldValue)
	typedNewValue, err2 := castBridgeDomainValue(key, newValue)
	if err1 != nil || err2 != nil {
		return false
	}
	return da.descriptor.ValueComparator(key, typedOldValue, typedNewValue)
}

func (da *BridgeDomainDescriptorAdapter) Add(key string, value proto.Message) (metadata Metadata, err error) {
	typedValue, err := castBridgeDomainValue(key, value)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Add(key, typedValue)
}

func (da *BridgeDomainDescriptorAdapter) Modify(key string, oldValue, newValue proto.Message, oldMetadata Metadata) (newMetadata Metadata, err error) {
	oldTypedValue, err := castBridgeDomainValue(key, oldValue)
	if err != nil {
		return nil, err
	}
	newTypedValue, err := castBridgeDomainValue(key, newValue)
	if err != nil {
		return nil, err
	}
	typedOldMetadata, err := castBridgeDomainMetadata(key, oldMetadata)
	if err != nil {
		return nil, err
	}
	return da.descriptor.Modify(key, oldTypedValue, newTypedValue, typedOldMetadata)
}

func (da *BridgeDomainDescriptorAdapter) Delete(key string, value proto.Message, metadata Metadata) error {
	typedValue, err := castBridgeDomainValue(key, value)
	if err != nil {
		return err
	}
	typedMetadata, err := castBridgeDomainMetadata(key, metadata)
	if err != nil {
		return err
	}
	return da.descriptor.Delete(key, typedValue, typedMetadata)
}

func (da *BridgeDomainDescriptorAdapter) ModifyWithRecreate(key string, oldValue, newValue proto.Message, metadata Metadata) bool {
	oldTypedValue, err := castBridgeDomainValue(key, oldValue)
	if err != nil {
		return true
	}
	newTypedValue, err := castBridgeDomainValue(key, newValue)
	if err != nil {
		return true
	}
	typedMetadata, err := castBridgeDomainMetadata(key, metadata)
	if err != nil {
		return true
	}
	return da.descriptor.ModifyWithRecreate(key, oldTypedValue, newTypedValue, typedMetadata)
}

func (da *BridgeDomainDescriptorAdapter) Update(key string, value proto.Message, metadata Metadata) error {
	typedValue, err := castBridgeDomainValue(key, value)
	if err != nil {
		return err
	}
	typedMetadata, err := castBridgeDomainMetadata(key, metadata)
	if err != nil {
		return err
	}
	return da.descriptor.Update(key, typedValue, typedMetadata)
}

func (da *BridgeDomainDescriptorAdapter) Dependencies(key string, value proto.Message) []Dependency {
	typedValue, err := castBridgeDomainValue(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.Dependencies(key, typedValue)
}

func (da *BridgeDomainDescriptorAdapter) DerivedValues(key string, value proto.Message) []KeyValuePair {
	typedValue, err := castBridgeDomainValue(key, value)
	if err != nil {
		return nil
	}
	return da.descriptor.DerivedValues(key, typedValue)
}

func (da *BridgeDomainDescriptorAdapter) Dump(correlate []KVWithMetadata) ([]KVWithMetadata, error) {
	var correlateWithType []BridgeDomainKVWithMetadata
	for _, kvpair := range correlate {
		typedValue, err := castBridgeDomainValue(kvpair.Key, kvpair.Value)
		if err != nil {
			continue
		}
		typedMetadata, err := castBridgeDomainMetadata(kvpair.Key, kvpair.Metadata)
		if err != nil {
			continue
		}
		correlateWithType = append(correlateWithType,
			BridgeDomainKVWithMetadata{
				Key:      kvpair.Key,
				Value:    typedValue,
				Metadata: typedMetadata,
				Origin:   kvpair.Origin,
			})
	}

	typedDump, err := da.descriptor.Dump(correlateWithType)
	if err != nil {
		return nil, err
	}
	var dump []KVWithMetadata
	for _, typedKVWithMetadata := range typedDump {
		kvWithMetadata := KVWithMetadata{
			Key:      typedKVWithMetadata.Key,
			Metadata: typedKVWithMetadata.Metadata,
			Origin:   typedKVWithMetadata.Origin,
		}
		kvWithMetadata.Value = typedKVWithMetadata.Value
		dump = append(dump, kvWithMetadata)
	}
	return dump, err
}

////////// Helper methods //////////

func castBridgeDomainValue(key string, value proto.Message) (*vpp_l2.BridgeDomain, error) {
	typedValue, ok := value.(*vpp_l2.BridgeDomain)
	if !ok {
		return nil, ErrInvalidValueType(key, value)
	}
	return typedValue, nil
}

func castBridgeDomainMetadata(key string, metadata Metadata) (*idxvpp2.OnlyIndex, error) {
	if metadata == nil {
		return nil, nil
	}
	typedMetadata, ok := metadata.(*idxvpp2.OnlyIndex)
	if !ok {
		return nil, ErrInvalidMetadataType(key)
	}
	return typedMetadata, nil
}
