# beautify and generate configuration file
cue fmt example/conf.cue
cue export example/conf.cue --out yaml > example/conf.yaml
echo "Done fmt cue and generate yaml configuration"
