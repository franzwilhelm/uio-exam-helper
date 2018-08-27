// Copyright © 2018 Franz von der Lippe franz.vonderlippe@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"github.com/franzwilhelm/uio-exam-helper/db/model"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	study               string
	faculty             string
	count               int
	abvailableFaculties = `uv, matnat, teologi, sv, odont, medisin, matnat, jus or hf`
)

// subjectCmd represents the subject command
var subjectCmd = &cobra.Command{
	Use:   "subject [subject code]",
	Short: "Get subject exam word count",
	Long:  abvailableFaculties,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Error("No subject specified")
			return
		}
		subjectID := args[0]
		log.Infof("Subject %s specified", subjectID)

		subject, err := model.GetSubjectByID(subjectID)
		if err != nil {
			subject, err = model.NewSubject(subjectID, faculty, study)
			if err != nil {
				log.WithError(err).Fatal("Could not create subject")
			}
		}
		if err := subject.FetchLatestResources(); err != nil {
			log.WithError(err).Fatal("Could not fetch latest resources")
		}
		if err := subject.Refresh(); err != nil {
			log.WithError(err).Fatal("Could not refresh subject")
		}
		subject.DownloadResources()
		strs, err := subject.GenerateWordTree()
		if err != nil {
			log.WithError(err).Fatal("Could not generate word tree")
		}
		for i := len(strs) - 1; i >= count; i-- {
			if len(strs[i]) > 0 {
				for _, s := range strs[i] {
					fmt.Println(s)
				}
			}
		}
	},
}

func init() {
	subjectCmd.Flags().IntVarP(&count, "count", "c", 3, "Count of word must be at least the number specified")
	subjectCmd.Flags().StringVarP(&faculty, "faculty", "f", "", abvailableFaculties)
	subjectCmd.Flags().StringVarP(&study, "study", "s", "", "Code of your study within the faculty")
	subjectCmd.MarkFlagRequired("faculty")
	subjectCmd.MarkFlagRequired("study")

	rootCmd.AddCommand(subjectCmd)
}
