# Fit dockerfile for goreleaser

for f in `ls build/docker`; do
  mkdir -p artifacts/docker/${f}
  cp -f "build/docker/${f}/Dockerfile" "artifacts/docker/${f}/Dockerfile"
  if ! which gsed 2>&1 >/dev/null; then
    # Linux
    sed -i 's#COPY artifacts/[a-z0-1]\+/#COPY #' "artifacts/docker/${f}/Dockerfile"
  else
    # MacOS
    gsed -i 's#COPY artifacts/[a-z0-1]\+/#COPY #' "artifacts/docker/${f}/Dockerfile"
  fi
done
