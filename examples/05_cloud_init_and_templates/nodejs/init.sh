#!/bin/sh

echo "Cloud init started for $HOSTNAME"

echo "01. Installing bash..."
apk add --no-cache --upgrade bash

echo "Cloud init finished for $HOSTNAME"