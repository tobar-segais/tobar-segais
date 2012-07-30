#!/bin/bash
if [ "A$(which pngcrush)A" == "AA" ]
then
  echo "pngcrush does not appear to be installed, or else it is not on your path. Please fix and try again"
  exit 1
fi
if [ "A$(which find)A" == "AA" ]
then
  echo "Fatal error: cannot find find on the PATH $PATH"
  exit 1
fi
for png in $(find $(dirname $0) -name \*.png | grep -v "/target/" )
do
  echo "Optimizing $png"
  rm -f "$png.crush"
  pngcrush -oldtimestamp -rem gAMA -rem cHRM -rem iCCP -rem sRGB -rem alla -rem text "$png" "$png.crush" >> /dev/null
  BEFORE=`ls -la "$png" | awk '{print $5}'`
  AFTER=`ls -la $png.crush | awk '{print $5}'`
  let REDUCED=$BEFORE-$AFTER
  if [ $AFTER -lt $BEFORE ]; then
    mv -f "$png.crush" "$png"
    echo "    saved $REDUCED bytes, file size is now $AFTER bytes"
  else
    rm -f "$png.crush"
  fi
done
