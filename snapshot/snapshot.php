#!/usr/bin/env php
<?php

define("DEFAULT_PORT", 8085);
define("DEFAULT_HEIGHT", 400);
define("DEFAULT_WIDTH", 500);
define("DEFAULT_FORMAT", "png");
define("DEFAULT_FROM", "10");
define("DEFAULT_PATH", "/data0/delong1");

$delay = 2;

function curl_request($url) {
    $ch = curl_init();
    curl_setopt($ch, CURLOPT_URL, $url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
    $res = curl_exec($ch);
    curl_close($ch);

    return $res;
}

function save_snapshot($path, $target, $content, $format, $delay) {
    $timestamp = time();
    $timestamp = $timestamp - $delay * 60;
    $date = date("Y-m-d", $timestamp);
    $hour = date("H", $timestamp);
    $min = date("i", $timestamp);

    if ($path[strlen($path)-1] == '/') {
        $path = substr($path, 0, strlen($path)-1);
    }  
    $path = $path . '/' . $date . '/' . $hour;
    if (!file_exists($path)) {
        mkdir($path, 0755, true);
    }
    $name = $target . "-" . $min . "." . $format;
    $file = $path . "/" . $name;
    file_put_contents($file, $content);
}

function _echo($str) {
    echo $str . "\n";
}

if ($argc < 2) {
    die("Usage: $argv[0] [config_ini]");
}
$ini_file = $argv[1];
$ini_array = parse_ini_file($ini_file, true);

if (array_key_exists("graphite", $ini_array)) {
    $graphite = $ini_array["graphite"];
    if (array_key_exists("IP", $graphite)) {
        $ip = $graphite["IP"];
    }
    if (array_key_exists("PORT", $graphite)) {
        $port = $graphite["PORT"];
    }
} else {
    die("Exception: no graphite info");
}

$targets = array();
if (array_key_exists("target", $ini_array)) {
    $target = $ini_array["target"];
    foreach($target as $key=>$value) {
        array_push($targets, $value);
    }
} else {
    die("Exception: no target info");
}

if (array_key_exists("imp", $ini_array)) {
    $imp = $ini_array["imp"];
    if (array_key_exists("HEIGHT", $imp)) {
        $height = $imp["HEIGHT"];
    }
    if (array_key_exists("WIDTH", $imp)) {
        $width = $imp["WIDTH"];
    }
    if (array_key_exists("FORMAT", $imp)) {
        $format = $imp["FORMAT"];
    }
    if (array_key_exists("PATH", $imp)) {
        $path = $imp["PATH"];
    }
}

if (array_key_exists("time", $ini_array)) {
    $time = $ini_array["time"];
    if (array_key_exists("FROM", $time)) {
        $from = $time["FROM"];
    }
}

if (empty($targets) || !isset($ip)) {
    die("No dest graphite or targets");
}
if (!isset($port)) {
    $port = DEFAULT_PORT;
}
if (!isset($height)) {
    $height = DEFAULT_HEIGHT;
}
if (!isset($width)) {
    $width = DEFAULT_WIDTH;
}
if (!isset($format)) {
    $format = DEFAULT_FORMAT;
}
if (!isset($path)) {
    $path = DEFAULT_PATH;
}
if (!isset($from)) {
    $from = DEFAULT_FROM;
}

$interval = $from * 60;
$from = $from + $delay;

while (true) {
    foreach ($targets as $target) {
        $url = "http://" . $ip . ":" . $port . "/render?";
        $params = array();
        $params["target"] = $target;
        $params["height"] = $height;
        $params["width"] = $width;
        $params["from"] = "-" . $from . "min";
        $params["until"] = "-" . $delay . "min";
        $params["title"] = $target;
        $params["format"] = $format;
        foreach ($params as $k=>$v) {
            $url .= $k . '=' . urlencode($v) . '&';
        }
        _echo($url);
        $content = curl_request($url);
        save_snapshot($path, $target, $content, $format, $delay);
    }
    sleep($interval);
}

?>
