# Update version tag of Makefile

if [ "${1}" != "" ]; then
  sed -i "/^TAG/s/dev/${1}/"
else
  echo "TAG should not be empty"
fi
