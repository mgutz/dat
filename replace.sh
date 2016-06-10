find . -name "*.go" -not -path "./node_modules/*" -print | xargs sed -i 's|"gopkg.in/mgutz/dat.v2"|"gopkg.in/mgutz/dat.v2/dat"|g'


