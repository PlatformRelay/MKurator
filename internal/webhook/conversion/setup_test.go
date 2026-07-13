package conversion

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("conversion webhook setup", func() {
	It("registers a conversion webhook for every hub kind", func() {
		mgr, err := newConversionManager(scheme.Scheme)
		Expect(err).NotTo(HaveOccurred())

		Expect(SetupWithManager(mgr)).To(Succeed())
	})

	It("wraps the failing kind when a hub type is missing from the scheme", func() {
		// An empty scheme has no v1beta1 hub types registered, so the first loop
		// entry (Queue) fails to resolve its GVK and the wrapped error must name
		// that kind.
		mgr, err := newConversionManager(runtime.NewScheme())
		Expect(err).NotTo(HaveOccurred())

		err = SetupWithManager(mgr)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("setup Queue conversion webhook"))
	})
})
