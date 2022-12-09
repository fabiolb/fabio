#!/usr/bin/env perl

use strict;

local $/;
$_ = <>;
if (/^### \[$ENV{RELEASE}.*?\n\s*(.*?)^### \[v/gms) {
    print $1;
}
