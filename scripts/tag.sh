# Create and push tag
log_info "Creating and pushing tag: $VERSION"
git tag -a "$VERSION" -m "Release $VERSION"

log_info "Pushing tag to origin..."
git push origin "$VERSION"