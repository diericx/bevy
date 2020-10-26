#!/bin/bash

# Recreate config file
rm -rf $1
touch $1

# Add assignment
echo "window._env_ = {" >> $1

if [ ! -z "${REACT_APP_TMDB_API_KEY}" ]; then
  echo "  REACT_APP_TMDB_API_KEY: \"$REACT_APP_TMDB_API_KEY\"," >> $1
fi
if [ ! -z "${BACKEND_URL}" ]; then
  echo "  BACKEND_URL: \"$BACKEND_URL\"," >> $1
fi

echo "}" >> $1
