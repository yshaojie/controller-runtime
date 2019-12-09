/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package envtest

import (
	"context"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Test", func() {
	var crds []*v1beta1.CustomResourceDefinition
	var err error
	var s *runtime.Scheme
	var c client.Client

	var validDirectory = filepath.Join(".", "testdata")
	var invalidDirectory = "fake"

	// Initialize the client
	BeforeEach(func(done Done) {
		crds = []*v1beta1.CustomResourceDefinition{}
		s = runtime.NewScheme()
		err = v1beta1.AddToScheme(s)
		Expect(err).NotTo(HaveOccurred())

		c, err = client.New(env.Config, client.Options{Scheme: s})
		Expect(err).NotTo(HaveOccurred())

		close(done)
	})

	// Cleanup CRDs
	AfterEach(func(done Done) {
		for _, crd := range crds {
			// Delete only if CRD exists.
			crdObjectKey := client.ObjectKey{
				Name: crd.Name,
			}
			var placeholder v1beta1.CustomResourceDefinition
			err := c.Get(context.TODO(), crdObjectKey, &placeholder)
			if err != nil && apierrors.IsNotFound(err) {
				// CRD doesn't need to be deleted.
				continue
			}
			Expect(err).NotTo(HaveOccurred())
			Expect(c.Delete(context.TODO(), crd)).To(Succeed())
		}
		close(done)
	})

	Describe("InstallCRDs", func() {
		It("should install the CRDs into the cluster using directory", func(done Done) {
			crds, err = InstallCRDs(env.Config, CRDInstallOptions{
				Paths: []string{validDirectory},
			})
			Expect(err).NotTo(HaveOccurred())

			// Expect to find the CRDs

			crd := &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "foos.bar.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("Foo"))

			crd = &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "bazs.qux.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("Baz"))

			crd = &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "captains.crew.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("Captain"))

			crd = &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "firstmates.crew.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("FirstMate"))

			crd = &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "drivers.crew.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("Driver"))

			err = WaitForCRDs(env.Config, []*v1beta1.CustomResourceDefinition{
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "qux.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "bazs",
						}},
				},
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "bar.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "foos",
						}},
				},
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "crew.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "captains",
						}},
				},
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "crew.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "firstmates",
						}},
				},
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group: "crew.example.com",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "drivers",
						},
						Versions: []v1beta1.CustomResourceDefinitionVersion{
							{
								Name:    "v1",
								Storage: true,
								Served:  true,
							},
							{
								Name:    "v2",
								Storage: false,
								Served:  true,
							},
						}},
				},
			},
				CRDInstallOptions{MaxTime: 50 * time.Millisecond, PollInterval: 15 * time.Millisecond},
			)
			Expect(err).NotTo(HaveOccurred())

			close(done)
		}, 5)

		It("should install the CRDs into the cluster using file", func(done Done) {
			crds, err = InstallCRDs(env.Config, CRDInstallOptions{
				Paths: []string{filepath.Join(".", "testdata", "examplecrd2.yaml")},
			})
			Expect(err).NotTo(HaveOccurred())

			crd := &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "bazs.qux.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("Baz"))

			err = WaitForCRDs(env.Config, []*v1beta1.CustomResourceDefinition{
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "qux.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "bazs",
						}},
				},
			},
				CRDInstallOptions{MaxTime: 50 * time.Millisecond, PollInterval: 15 * time.Millisecond},
			)
			Expect(err).NotTo(HaveOccurred())

			close(done)
		}, 10)

		It("should filter out already existent CRD", func(done Done) {
			crds, err = InstallCRDs(env.Config, CRDInstallOptions{
				Paths: []string{
					filepath.Join(".", "testdata"),
					filepath.Join(".", "testdata", "examplecrd1.yaml"),
				},
			})
			Expect(err).NotTo(HaveOccurred())

			crd := &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "foos.bar.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("Foo"))

			err = WaitForCRDs(env.Config, []*v1beta1.CustomResourceDefinition{
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "bar.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "foos",
						}},
				},
			},
				CRDInstallOptions{MaxTime: 50 * time.Millisecond, PollInterval: 15 * time.Millisecond},
			)
			Expect(err).NotTo(HaveOccurred())

			close(done)
		}, 10)

		It("should not return an not error if the directory doesn't exist", func(done Done) {
			crds, err = InstallCRDs(env.Config, CRDInstallOptions{Paths: []string{invalidDirectory}})
			Expect(err).NotTo(HaveOccurred())

			close(done)
		}, 5)

		It("should return an error if the directory doesn't exist", func(done Done) {
			crds, err = InstallCRDs(env.Config, CRDInstallOptions{
				Paths: []string{invalidDirectory}, ErrorIfPathMissing: true,
			})
			Expect(err).To(HaveOccurred())

			close(done)
		}, 5)

		It("should return an error if the file doesn't exist", func(done Done) {
			crds, err = InstallCRDs(env.Config, CRDInstallOptions{Paths: []string{
				filepath.Join(".", "testdata", "fake.yaml")}, ErrorIfPathMissing: true,
			})
			Expect(err).To(HaveOccurred())

			close(done)
		}, 5)

		It("should return an error if the resource group version isn't found", func(done Done) {
			// Wait for a CRD where the Group and Version don't exist
			err := WaitForCRDs(env.Config,
				[]*v1beta1.CustomResourceDefinition{
					{
						Spec: v1beta1.CustomResourceDefinitionSpec{
							Version: "v1",
							Names: v1beta1.CustomResourceDefinitionNames{
								Plural: "notfound",
							}},
					},
				},
				CRDInstallOptions{MaxTime: 50 * time.Millisecond, PollInterval: 15 * time.Millisecond},
			)
			Expect(err).To(HaveOccurred())

			close(done)
		}, 5)

		It("should return an error if the resource isn't found in the group version", func(done Done) {
			crds, err = InstallCRDs(env.Config, CRDInstallOptions{
				Paths: []string{"."},
			})
			Expect(err).NotTo(HaveOccurred())

			// Wait for a CRD that doesn't exist, but the Group and Version do
			err = WaitForCRDs(env.Config, []*v1beta1.CustomResourceDefinition{
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "qux.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "bazs",
						}},
				},
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "bar.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "fake",
						}},
				}},
				CRDInstallOptions{MaxTime: 50 * time.Millisecond, PollInterval: 15 * time.Millisecond},
			)
			Expect(err).To(HaveOccurred())

			close(done)
		}, 5)
	})

	Describe("UninstallCRDs", func() {
		It("should uninstall the CRDs from the cluster", func(done Done) {

			crds, err = InstallCRDs(env.Config, CRDInstallOptions{
				Paths: []string{validDirectory},
			})
			Expect(err).NotTo(HaveOccurred())

			// Expect to find the CRDs

			crd := &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "foos.bar.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("Foo"))

			crd = &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "bazs.qux.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("Baz"))

			crd = &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "captains.crew.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("Captain"))

			crd = &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "firstmates.crew.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("FirstMate"))

			crd = &v1beta1.CustomResourceDefinition{}
			err = c.Get(context.TODO(), types.NamespacedName{Name: "drivers.crew.example.com"}, crd)
			Expect(err).NotTo(HaveOccurred())
			Expect(crd.Spec.Names.Kind).To(Equal("Driver"))

			err = WaitForCRDs(env.Config, []*v1beta1.CustomResourceDefinition{
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "qux.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "bazs",
						}},
				},
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "bar.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "foos",
						}},
				},
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "crew.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "captains",
						}},
				},
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group:   "crew.example.com",
						Version: "v1beta1",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "firstmates",
						}},
				},
				{
					Spec: v1beta1.CustomResourceDefinitionSpec{
						Group: "crew.example.com",
						Names: v1beta1.CustomResourceDefinitionNames{
							Plural: "drivers",
						},
						Versions: []v1beta1.CustomResourceDefinitionVersion{
							{
								Name:    "v1",
								Storage: true,
								Served:  true,
							},
							{
								Name:    "v2",
								Storage: false,
								Served:  true,
							},
						}},
				},
			},
				CRDInstallOptions{MaxTime: 50 * time.Millisecond, PollInterval: 15 * time.Millisecond},
			)
			Expect(err).NotTo(HaveOccurred())

			err = UninstallCRDs(env.Config, CRDInstallOptions{
				Paths: []string{validDirectory},
			})
			Expect(err).NotTo(HaveOccurred())

			// Expect to NOT find the CRDs

			crds := []string{
				"foos.bar.example.com",
				"bazs.qux.example.com",
				"captains.crew.example.com",
				"firstmates.crew.example.com",
				"drivers.crew.example.com",
			}
			placeholder := &v1beta1.CustomResourceDefinition{}
			Eventually(func() bool {
				for _, crd := range crds {
					err = c.Get(context.TODO(), types.NamespacedName{Name: crd}, placeholder)
					notFound := err != nil && apierrors.IsNotFound(err)
					if !notFound {
						return false
					}
				}
				return true
			}, 20).Should(BeTrue())

			close(done)
		}, 30)
	})

	Describe("Start", func() {
		It("should raise an error on invalid dir when flag is enabled", func(done Done) {
			env = &Environment{ErrorIfCRDPathMissing: true, CRDDirectoryPaths: []string{invalidDirectory}}
			_, err := env.Start()
			Expect(err).To(HaveOccurred())
			close(done)
		}, 30)

		It("should not raise an error on invalid dir when flag is disabled", func(done Done) {
			env = &Environment{ErrorIfCRDPathMissing: false, CRDDirectoryPaths: []string{invalidDirectory}}
			_, err := env.Start()
			Expect(err).NotTo(HaveOccurred())
			close(done)
		}, 30)
	})
})
