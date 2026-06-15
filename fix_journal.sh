# Restore previous sentinel.md if possible and append instead. Wait, it might be too late to git restore if we didn't commit the original, but let's check git status.
git diff HEAD .jules/sentinel.md
