package localrunner

import (
	"path"

	"github.com/profMagija/cloudpost/config"
)

func init_bucket(flock *config.Flock, f *config.Bucket, s map[string]*storageObject) {
	for _, file := range f.Contents {
		pp := path.Join(flock.Root, file.Src)

		mimeType := file.Type

		if mimeType == "" {
			mimeType = "text/plain"
		}

		s[file.Dst] = &storageObject{
			onDiskPath: &pp,
			mimeType:   mimeType,
		}

		local_log_verbose("[%s] added %s from %s as %s", f.Name, file.Dst, file.Src, mimeType)
	}
}
