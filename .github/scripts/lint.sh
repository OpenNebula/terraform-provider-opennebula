#!/bin/bash

export FMT_FILES="$(go fmt ./...)"
if [ -n "${FMT_FILES}" ]; then
  echo "\n🥲 Oops!\nPlease formate following files with go fmt:"
  echo "\n${FMT_FILES}"
  echo "# 🥲 Oops!" >> $GITHUB_STEP_SUMMARY
  echo "Please formate the following files with \`go fmt\`:\n" >> $GITHUB_STEP_SUMMARY
  echo "\`\`\`" >> $GITHUB_STEP_SUMMARY
  echo "${FMT_FILES}" >> $GITHUB_STEP_SUMMARY
  echo "\`\`\`" >> $GITHUB_STEP_SUMMARY
  exit $(echo "${FMT_FILES}" | grep wc -l)
  echo ""

else
  echo "\n🎉 All good!"
  echo "# 🎉 All good!" >> $GITHUB_STEP_SUMMARY
  echo "All files are formated as expected!" >> $GITHUB_STEP_SUMMARY
fi

exit 0
