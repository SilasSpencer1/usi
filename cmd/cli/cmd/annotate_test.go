package cmd_test

import (
	"errors"
	"fmt"

	"platform-go-common/pkg/util"

	workspacemocks "/usi/pkg/workspace/mocks"
	"usi/pkg/client"

	"usi/pkg/core"

	"usi/cmd/cli/cmd"
	"usi/cmd/cli/cmd/mocks"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("Annotate", func() {
	var (
		mockCtrl *gomock.Controller
		wsMock   *workspacemocks.MockWorkspace

		name        = "test-resource"
		uuid        = util.StrPtr("011dbd3e-a3d0-11ed-a8fc-0242ac120002")
		annotations = util.StrPtr("key1=value1,key2=value2")

		action = func() {
			cmd.AnnotateAction(wsMock, uuid, annotations)
		}
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		wsMock = workspacemocks.NewMockWorkspace(mockCtrl)
		stubMetricsReporter(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Context("When the registry returns data", func() {
		It("should print the resource", func() {
			returnedResource := &client.Resource{
				UUID:     *uuid,
				Name:     name,
				TypeName: util.StrPtr("type"),
			}
			wsMock.EXPECT().Annotate(
				gomock.Any(),
				core.Request{UUID: uuid},
				gomock.Any(),
			).Return(returnedResource, nil)
			results, err := captureStdOutErrAsync(action)
			Expect(err).To(BeNil())
			Eventually(results.StdOutBuffer).Should(gbytes.Say(fmt.Sprintf("type > %s", name)))
			Eventually(results.StdOutBuffer).Should(gbytes.Say(fmt.Sprintf("name: %s", name)))
			Eventually(results.StdOutBuffer).Should(gbytes.Say(fmt.Sprintf("uuid: %s", *uuid)))
		})
	})

	Context("When the registry returns an error", func() {
		It("should handle the error as expected", func() {
			errMsg := "Test error"
			wsMock.EXPECT().Annotate(
				gomock.Any(),
				core.Request{UUID: uuid},
				gomock.Any(),
			).Return(nil, errors.New(errMsg))

			reporterMock := mocks.NewMockUsageReporter(mockCtrl)
			cmd.Reporter = reporterMock
			reporterMock.EXPECT().SendHoneycombEvent(
				"annotate",
				map[string]interface{}{"error_type": "", "result": "failure", "remote": "false"},
			)
			reporterMock.EXPECT().FlushHoneycomb()

			results, err := captureStdOutErrAsync(action)
			Expect(err).To(BeNil())
			Eventually(results.StdErrBuffer).Should(gbytes.Say(errMsg))
			Eventually(*results.ExitedOnErr).Should(BeTrue())
		})
	})

	Context("When annotating with addl_environments annotations format", func() {
		var (
			expectedAnnotations = "addl_annotations=testing>na;testing>eu"
		)
		It("should pass correct addl_annotations to the resource", func() {

			action = func() {
				cmd.AnnotateAction(wsMock, uuid, &expectedAnnotations)
			}

			returnedResource := &client.Resource{
				UUID:     *uuid,
				Name:     name,
				TypeName: util.StrPtr("type"),
			}
			wsMock.EXPECT().Annotate(
				gomock.Any(),
				core.Request{UUID: uuid},
				gomock.Any(),
			).Return(returnedResource, nil)
			results, err := captureStdOutErrAsync(action)
			Expect(err).To(BeNil())
			Eventually(results.StdOutBuffer).Should(gbytes.Say(fmt.Sprintf("type > %s", name)))
			Eventually(results.StdOutBuffer).Should(gbytes.Say(fmt.Sprintf("name: %s", name)))
			Eventually(results.StdOutBuffer).Should(gbytes.Say(fmt.Sprintf("uuid: %s", *uuid)))
		})
	})

})
