//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1beta

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AddressEndpoint) DeepCopyInto(out *AddressEndpoint) {
	*out = *in
	if in.ToPorts != nil {
		in, out := &in.ToPorts, &out.ToPorts
		*out = make([]PortInfo, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AddressEndpoint.
func (in *AddressEndpoint) DeepCopy() *AddressEndpoint {
	if in == nil {
		return nil
	}
	out := new(AddressEndpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BalancingBackend) DeepCopyInto(out *BalancingBackend) {
	*out = *in
	if in.AddressEndpoint != nil {
		in, out := &in.AddressEndpoint, &out.AddressEndpoint
		*out = make([]*AddressEndpoint, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(AddressEndpoint)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.ServiceEndpoint != nil {
		in, out := &in.ServiceEndpoint, &out.ServiceEndpoint
		*out = new(ServiceEndpoint)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BalancingBackend.
func (in *BalancingBackend) DeepCopy() *BalancingBackend {
	if in == nil {
		return nil
	}
	out := new(BalancingBackend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BalancingPolicy) DeepCopyInto(out *BalancingPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BalancingPolicy.
func (in *BalancingPolicy) DeepCopy() *BalancingPolicy {
	if in == nil {
		return nil
	}
	out := new(BalancingPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BalancingPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BalancingPolicyList) DeepCopyInto(out *BalancingPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]BalancingPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BalancingPolicyList.
func (in *BalancingPolicyList) DeepCopy() *BalancingPolicyList {
	if in == nil {
		return nil
	}
	out := new(BalancingPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BalancingPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BalancingSpec) DeepCopyInto(out *BalancingSpec) {
	*out = *in
	if in.Enabled != nil {
		in, out := &in.Enabled, &out.Enabled
		*out = new(bool)
		**out = **in
	}
	in.BalancingFrontend.DeepCopyInto(&out.BalancingFrontend)
	in.BalancingBackend.DeepCopyInto(&out.BalancingBackend)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BalancingSpec.
func (in *BalancingSpec) DeepCopy() *BalancingSpec {
	if in == nil {
		return nil
	}
	out := new(BalancingSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BalancingStatus) DeepCopyInto(out *BalancingStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BalancingStatus.
func (in *BalancingStatus) DeepCopy() *BalancingStatus {
	if in == nil {
		return nil
	}
	out := new(BalancingStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalRedirectBackend) DeepCopyInto(out *LocalRedirectBackend) {
	*out = *in
	in.LocalEndpointSelector.DeepCopyInto(&out.LocalEndpointSelector)
	if in.ToPorts != nil {
		in, out := &in.ToPorts, &out.ToPorts
		*out = make([]PortInfo, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalRedirectBackend.
func (in *LocalRedirectBackend) DeepCopy() *LocalRedirectBackend {
	if in == nil {
		return nil
	}
	out := new(LocalRedirectBackend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalRedirectPolicy) DeepCopyInto(out *LocalRedirectPolicy) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalRedirectPolicy.
func (in *LocalRedirectPolicy) DeepCopy() *LocalRedirectPolicy {
	if in == nil {
		return nil
	}
	out := new(LocalRedirectPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LocalRedirectPolicy) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalRedirectPolicyList) DeepCopyInto(out *LocalRedirectPolicyList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]LocalRedirectPolicy, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalRedirectPolicyList.
func (in *LocalRedirectPolicyList) DeepCopy() *LocalRedirectPolicyList {
	if in == nil {
		return nil
	}
	out := new(LocalRedirectPolicyList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *LocalRedirectPolicyList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalRedirectSpec) DeepCopyInto(out *LocalRedirectSpec) {
	*out = *in
	if in.Enabled != nil {
		in, out := &in.Enabled, &out.Enabled
		*out = new(bool)
		**out = **in
	}
	in.RedirectFrontend.DeepCopyInto(&out.RedirectFrontend)
	in.LocalRedirectBackend.DeepCopyInto(&out.LocalRedirectBackend)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalRedirectSpec.
func (in *LocalRedirectSpec) DeepCopy() *LocalRedirectSpec {
	if in == nil {
		return nil
	}
	out := new(LocalRedirectSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *LocalRedirectStatus) DeepCopyInto(out *LocalRedirectStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new LocalRedirectStatus.
func (in *LocalRedirectStatus) DeepCopy() *LocalRedirectStatus {
	if in == nil {
		return nil
	}
	out := new(LocalRedirectStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PortInfo) DeepCopyInto(out *PortInfo) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PortInfo.
func (in *PortInfo) DeepCopy() *PortInfo {
	if in == nil {
		return nil
	}
	out := new(PortInfo)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RedirectFrontend) DeepCopyInto(out *RedirectFrontend) {
	*out = *in
	if in.AddressMatcher != nil {
		in, out := &in.AddressMatcher, &out.AddressMatcher
		*out = new(AddressEndpoint)
		(*in).DeepCopyInto(*out)
	}
	if in.ServiceMatcher != nil {
		in, out := &in.ServiceMatcher, &out.ServiceMatcher
		*out = new(ServiceMatcher)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RedirectFrontend.
func (in *RedirectFrontend) DeepCopy() *RedirectFrontend {
	if in == nil {
		return nil
	}
	out := new(RedirectFrontend)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceEndpoint) DeepCopyInto(out *ServiceEndpoint) {
	*out = *in
	if in.ToPorts != nil {
		in, out := &in.ToPorts, &out.ToPorts
		*out = make([]PortInfo, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceEndpoint.
func (in *ServiceEndpoint) DeepCopy() *ServiceEndpoint {
	if in == nil {
		return nil
	}
	out := new(ServiceEndpoint)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceMatcher) DeepCopyInto(out *ServiceMatcher) {
	*out = *in
	if in.ToPorts != nil {
		in, out := &in.ToPorts, &out.ToPorts
		*out = make([]PortInfo, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceMatcher.
func (in *ServiceMatcher) DeepCopy() *ServiceMatcher {
	if in == nil {
		return nil
	}
	out := new(ServiceMatcher)
	in.DeepCopyInto(out)
	return out
}
