package modrinth

import "orchestrator/external/mm"

func mapHitToSummary(h Hit) mm.ContentSummary {
	return mm.ContentSummary{
		ID:            h.ProjectID,
		Slug:          h.Slug,
		Title:         h.Title,
		Description:   h.Description,
		IconURL:       h.IconURL,
		Author:        h.Author,
		Downloads:     h.Downloads,
		ServerSide:    h.ServerSide,
		ClientSide:    h.ClientSide,
		Versions:      h.Versions,
		LatestVersion: h.LatestVersion,
	}
}

func mapProjectToDetail(p Project) *mm.ContentDetail {
	return &mm.ContentDetail{
		ID:          p.ID,
		Slug:        p.Slug,
		Title:       p.Title,
		Description: p.Description,
		IconURL:     p.IconURL,
		Author:      p.Author,
		Downloads:   p.Downloads,
		ServerSide:  p.ServerSide,
		ClientSide:  p.ClientSide,
	}
}

func mapVersion(v Version) mm.ContentVersion {
	files := make([]mm.ContentFile, len(v.Files))
	for i, f := range v.Files {
		files[i] = mm.ContentFile{
			URL:      f.URL,
			Filename: f.Filename,
			Primary:  f.Primary,
		}
	}
	deps := make([]mm.ContentDependency, len(v.Dependencies))
	for i, d := range v.Dependencies {
		deps[i] = mm.ContentDependency{
			VersionID:      d.VersionID,
			ProjectID:      d.ProjectID,
			DependencyType: d.DependencyType,
		}
	}
	return mm.ContentVersion{
		ID:            v.ID,
		ProjectID:     v.ProjectID,
		VersionNumber: v.VersionNumber,
		GameVersions:  v.GameVersions,
		Loaders:       v.Loaders,
		Files:         files,
		Dependencies:  deps,
	}
}
