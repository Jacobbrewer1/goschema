#!/bin/bash

function fail() {
  gum style --foreground 196 "$1"
  exit 1
}

if ! command -v gum &> /dev/null; then
  echo "gum is required to generate models.'"
  echo "Trying to install gum..."
  go install github.com/charmbracelet/gum@latest || fail "Failed to install gum"
fi

if ! command -v goschema &> /dev/null; then
  gum style --foreground 196 "goschema is required to generate models. Please install it"
  exit 1
fi

if ! command -v goimports &> /dev/null; then
  gum style --foreground 196 "goimports is required to generate models. Please install it by running 'go get golang.org/x/tools/cmd/goimports'"
  exit 1
fi

up=false
down=false
forced=false

# Get the flags passed to the script and set the variables accordingly
while getopts "udcf" flag; do
  case $flag in
    u)
      up=true
      ;;
    d)
      down=true
      ;;
    f)
      forced=true
      ;;
    *)
      gum style --foreground 196 "Invalid flag $flag"
      exit 1
      ;;
  esac
done

option=""

if [ "$up" = false ] && [ "$down" = false ]; then
  # Assume being ran by the user
  gum style --foreground 222 "Please choose an option"
  option=$(gum choose "create" "up" "down")
fi

if [ "$up" = true ] && [ "$down" = true ]; then
  gum style --foreground 196 "Cannot run both up and down migrations"
  exit 1
fi

case $option in
  up)
    # Is the DATABASE_URL set?
    if [ -z "$DATABASE_URL" ]; then
      gum style --foreground 196 "DATABASE_URL is not set"
      exit 1
    fi

    gum spin --spinner dot --title "Running up migrations" -- goschema migrate --up --loc=./migrations
    ;;
  down)
    # Is the DATABASE_URL set?
    if [ -z "$DATABASE_URL" ]; then
      gum style --foreground 196 "DATABASE_URL is not set"
      exit 1
    fi

    gum spin --spinner dot --title "Running down migrations" -- goschema migrate --down --loc=./migrations
    ;;
  create)
    name=$(gum input --placeholder "Please describe the migration")
    gum spin --spinner dot --title "Creating migrations" -- goschema create --out=./migrations --name="$name"
esac
