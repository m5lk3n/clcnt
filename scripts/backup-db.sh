if [[ -d ~/storage/downloads ]]; then
  cp ../clcnt.db ~/storage/downloads
  echo "DB backed up"
else
  echo "DB NOT backed up!"
fi
