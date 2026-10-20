package seqtoid

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/IT-Academic-Research-Services/seqtoid-cli/pkg/upload"
)

func UploadSamplesFlow(
	sampleFiles map[string]SampleFiles,
	stringMetadata map[string]string,
	projectName string,
	metadataCSVPath string,
	workflow string,
	workflowVersion string,
	sampleOptions SampleOptions,
	disableBuffer bool,
	resume bool,
) error {
	if workflowVersion != "" {
		if err := DefaultClient.ValidateWorkflowVersion(workflow, workflowVersion); err != nil {
			log.Fatal(err)
		}
	}

	projectID, err := DefaultClient.GetProjectID(projectName)
	if err != nil {
		log.Fatal(err)
	}

	// --resume: finish the samples an earlier run of this command created but never uploaded, instead
	// of creating new samples (which would duplicate or clash with them). Metadata was saved when the
	// samples were created, so none of it is re-sent or re-validated.
	if resume {
		names := make([]string, 0, len(sampleFiles))
		for name := range sampleFiles {
			names = append(names, name)
		}
		sort.Strings(names)
		samples, skipped, err := DefaultClient.ResumeUploadTargets(projectID, names)
		if err != nil {
			log.Fatal(err)
		}
		for _, name := range names {
			if reason, ok := skipped[name]; ok {
				fmt.Printf("skipping sample %q: %s\n", name, reason)
			}
		}
		if len(samples) == 0 {
			return fmt.Errorf("nothing to resume: none of the %d sample(s) in this directory are waiting to be uploaded to project %q", len(names), projectName)
		}
		fmt.Printf("resuming %d sample(s) in project %q\n", len(samples), projectName)
		return uploadAll(samples, sampleFiles, disableBuffer, projectName)
	}

	samplesMetadata, err := GetCombinedMetadata(sampleFiles, stringMetadata, metadataCSVPath)
	if err != nil {
		log.Fatal(err)
	}

	sampleNames := make([]string, 0, len(sampleFiles))
	for sampleName := range samplesMetadata {
		sampleNames = append(sampleNames, sampleName)
	}
	newSampleNames, err := DefaultClient.ValidateSampleNames(sampleNames, projectID)
	if err != nil {
		log.Fatal(err)
	}
	if len(sampleNames) != len(newSampleNames) {
		log.Fatal("error validating sample names")
	}
	for i := range sampleNames {
		if newSampleNames[i] != sampleNames[i] {
			samplesMetadata[newSampleNames[i]] = samplesMetadata[sampleNames[i]]
			delete(samplesMetadata, sampleNames[i])
			sampleFiles[newSampleNames[i]] = sampleFiles[sampleNames[i]]
			delete(sampleFiles, sampleNames[i])
		}
	}

	err = GeoSearchSuggestions(&samplesMetadata)
	if err != nil {
		log.Fatal(err)
	}
	err = DefaultClient.ValidateSamplesMetadata(projectID, samplesMetadata)
	if err != nil {
		if err.Error() == "metadata validation failed" {
			os.Exit(1)
		}
		log.Fatal(err)
	}

	samples, err := DefaultClient.CreateSamples(
		projectID,
		sampleFiles,
		samplesMetadata,
		workflow,
		workflowVersion,
		sampleOptions,
	)
	if err != nil {
		log.Fatal(err)
	}

	return uploadAll(samples, sampleFiles, disableBuffer, projectName)
}

func uploadSample(sample createSamplesResSample, sampleFiles map[string]SampleFiles, disableBuffer bool) error {
	credentials, err := DefaultClient.GetUploadCredentials(sample.ID)
	if err != nil {
		return err
	}
	u := upload.NewUploader(credentials, disableBuffer)
	sF := sampleFiles[sample.Name]
	for _, inputFile := range pendingInputFiles(sample) {
		// The server obfuscates the stored object name (inputFile.S3Path) so its basename can
		// no longer be matched to a local file. Correlate on inputFile.Source (the original
		// name the client sent) instead, falling back to the S3Path basename for older servers
		// that do not return Source. Files still upload to inputFile.S3Path (the real target).
		correlationName := filepath.Base(inputFile.S3Path)
		if inputFile.Source != "" {
			correlationName = filepath.Base(inputFile.Source)
		}
		var filenames []string
		if len(sF.R1) > 0 && filepath.Base(StripLaneNumber(sF.R1[0])) == correlationName {
			filenames = sF.R1
		} else if len(sF.R2) > 0 && filepath.Base(StripLaneNumber(sF.R2[0])) == correlationName {
			filenames = sF.R2
		} else if len(sF.Single) > 0 && filepath.Base(sF.Single[0]) == correlationName {
			filenames = sF.Single
		} else if len(sF.ReferenceFasta) > 0 && filepath.Base(sF.ReferenceFasta[0]) == correlationName {
			filenames = sF.ReferenceFasta
		} else if len(sF.PrimerBed) > 0 && filepath.Base(sF.PrimerBed[0]) == correlationName {
			filenames = sF.PrimerBed
		} else {
			allFilenames := []string{}
			if len(sF.R1) > 0 {
				allFilenames = append(allFilenames, StripLaneNumber(sF.R1[0]))
			}
			if len(sF.R2) > 0 {
				allFilenames = append(allFilenames, StripLaneNumber(sF.R2[0]))
			}
			if len(sF.Single) > 0 {
				allFilenames = append(allFilenames, StripLaneNumber(sF.Single[0]))
			}
			if len(sF.ReferenceFasta) > 0 {
				allFilenames = append(allFilenames, StripLaneNumber(sF.ReferenceFasta[0]))
			}
			if len(sF.PrimerBed) > 0 {
				allFilenames = append(allFilenames, StripLaneNumber(sF.PrimerBed[0]))
			}

			return fmt.Errorf("could not correlate server input file (source %q, s3 path %s) with any local file: %s", inputFile.Source, inputFile.S3Path, strings.Join(allFilenames, ", "))
		}
		if err := u.UploadFiles(filenames, inputFile.S3Path, inputFile.MultipartUploadId); err != nil {
			return err
		}
	}
	return DefaultClient.MarkSampleUploaded(sample.ID, sample.Name)
}

// uploadAll uploads each sample in turn. One sample failing (a dropped connection, a stalled network)
// no longer ends the whole batch: the rest still upload, and the failures are listed at the end with
// the command that finishes them.
func uploadAll(samples []createSamplesResSample, sampleFiles map[string]SampleFiles, disableBuffer bool, projectName string) error {
	var failed []string
	for _, sample := range samples {
		if err := uploadSample(sample, sampleFiles, disableBuffer); err != nil {
			fmt.Printf("failed to upload sample %q: %v\n", sample.Name, err)
			failed = append(failed, sample.Name)
			continue
		}
		fmt.Printf("uploaded sample %q\n", sample.Name)
	}
	if len(failed) > 0 {
		fmt.Printf("\n%d of %d sample(s) did not finish uploading: %s\n", len(failed), len(samples), strings.Join(failed, ", "))
		fmt.Println("Re-run the same command with --resume to finish them (nothing already uploaded is sent again).")
		return fmt.Errorf("%d sample(s) failed to upload", len(failed))
	}
	fmt.Printf("done: uploaded %d sample(s) to project %q; they are now processing on SeqToID\n", len(samples), projectName)
	return nil
}
