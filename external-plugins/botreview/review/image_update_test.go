/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright the KubeVirt authors.
 *
 */

package review

import (
	"github.com/sourcegraph/go-diff/diff"
	"os"
	"reflect"
	"testing"
)

func TestProwJobImageUpdate_Review(t1 *testing.T) {
	diffFilePaths := []string{
		"testdata/simple_bump-prow-job-images_sh.patch0",
		"testdata/simple_bump-prow-job-images_sh.patch1",
		"testdata/mixed_bump_prow_job.patch0",
	}
	diffFilePathsToDiffs := map[string]*diff.FileDiff{}
	for _, diffFile := range diffFilePaths {
		bumpImagesDiffFile, err := os.ReadFile(diffFile)
		if err != nil {
			t1.Errorf("failed to read diff: %v", err)
		}
		bumpFileDiffs, err := diff.ParseFileDiff(bumpImagesDiffFile)
		if err != nil {
			t1.Errorf("failed to read diff: %v", err)
		}
		diffFilePathsToDiffs[diffFile] = bumpFileDiffs
	}
	type fields struct {
		relevantFileDiffs []*diff.FileDiff
	}
	tests := []struct {
		name   string
		fields fields
		want   *ProwJobImageUpdateResult
	}{
		{
			name: "simple image bump",
			fields: fields{
				relevantFileDiffs: []*diff.FileDiff{
					diffFilePathsToDiffs["testdata/simple_bump-prow-job-images_sh.patch0"],
					diffFilePathsToDiffs["testdata/simple_bump-prow-job-images_sh.patch1"],
				},
			},
			want: &ProwJobImageUpdateResult{},
		},
		{
			name: "mixed image bump",
			fields: fields{
				relevantFileDiffs: []*diff.FileDiff{
					diffFilePathsToDiffs["testdata/mixed_bump_prow_job.patch0"],
				},
			},
			want: &ProwJobImageUpdateResult{
				notMatchingHunks: map[string][]*diff.Hunk{"github/ci/prow-deploy/files/jobs/kubevirt/kubevirt/kubevirt-presubmits.yaml": {diffFilePathsToDiffs["testdata/mixed_bump_prow_job.patch0"].Hunks[0]}},
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ProwJobImageUpdate{
				relevantFileDiffs: tt.fields.relevantFileDiffs,
			}
			if got := t.Review(); !reflect.DeepEqual(got, tt.want) {
				t1.Errorf("Review() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProwJobImageUpdate_AddIfRelevant(t1 *testing.T) {
	type fields struct {
		relevantFileDiffs []*diff.FileDiff
		notMatchingHunks  []*diff.Hunk
	}
	type args struct {
		fileDiff *diff.FileDiff
	}
	tests := []struct {
		name                      string
		fields                    fields
		args                      args
		expectedRelevantFileDiffs []*diff.FileDiff
	}{
		{
			name: "release branch config is ignored",
			fields: fields{
				relevantFileDiffs: nil,
				notMatchingHunks:  nil,
			},
			args: args{
				fileDiff: &diff.FileDiff{
					OrigName: "a/github/ci/prow-deploy/files/jobs/kubevirt/kubevirt/kubevirt-presubmits-0.54.yaml",
					OrigTime: nil,
					NewName:  "b/github/ci/prow-deploy/files/jobs/kubevirt/kubevirt/kubevirt-presubmits-0.54.yaml",
					NewTime:  nil,
					Extended: nil,
					Hunks:    nil,
				},
			},
		},
		{
			name: "non-release branch config is added",
			fields: fields{
				relevantFileDiffs: nil,
				notMatchingHunks:  nil,
			},
			args: args{
				fileDiff: &diff.FileDiff{
					OrigName: "a/github/ci/prow-deploy/files/jobs/kubevirt/kubevirt/kubevirt-presubmits.yaml",
					OrigTime: nil,
					NewName:  "b/github/ci/prow-deploy/files/jobs/kubevirt/kubevirt/kubevirt-presubmits.yaml",
					NewTime:  nil,
					Extended: nil,
					Hunks: []*diff.Hunk{
						{Body: []byte("+          - image: quay.io/kubevirtci/bootstrap:v20220110-c066ff5")},
					},
				},
			},
			expectedRelevantFileDiffs: []*diff.FileDiff{
				{
					OrigName: "a/github/ci/prow-deploy/files/jobs/kubevirt/kubevirt/kubevirt-presubmits.yaml",
					OrigTime: nil,
					NewName:  "b/github/ci/prow-deploy/files/jobs/kubevirt/kubevirt/kubevirt-presubmits.yaml",
					NewTime:  nil,
					Extended: nil,
					Hunks: []*diff.Hunk{
						{Body: []byte("+          - image: quay.io/kubevirtci/bootstrap:v20220110-c066ff5")},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &ProwJobImageUpdate{
				relevantFileDiffs: tt.fields.relevantFileDiffs,
				notMatchingHunks:  tt.fields.notMatchingHunks,
			}
			t.AddIfRelevant(tt.args.fileDiff)
			if !reflect.DeepEqual(tt.expectedRelevantFileDiffs, t.relevantFileDiffs) {
				t1.Errorf("expectedRelevantFileDiffs not equal: %v\n, was\n%v", tt.expectedRelevantFileDiffs, t.relevantFileDiffs)
			}
		})
	}
}
