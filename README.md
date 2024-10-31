# Firefly-iii webhooks

Webhooks handler for Firefly-iii

## Available configurations

Options can be handled as environment variables or flags. 
Use lower-kebab-case for flags

- ADDR network address and port to listen to. Defaults to ":4000"
- LOG_LEVEL log message levels to display. Defaults to "debug", "error", "warn", "info" and "debug" available
- FIREFLY_BASE_URL firefly-iii instance endpoint. **MUST NOT** end with / e.g. https://firefly.example.com
- FIREFLY_CONFIG json configuration files to use for webhooks. Defaults to ./config.json
- FIREFLY_API_KEY personal access token generated from Firefly-iii settings

The FIREFLY_CONFIG file must be a json object with keys the actions handled and values an array of configurations. 
Each configuration depends on the action.

## Available actions

### Split amount

Split a transaction updating the amount and foreign amount based on configuration and conditionally create a new transaction
with the same values but the source account and amount.

e.g. When I perform a withdrawal on account X calculate the integer division from the foreign amount updating the transaction
amount and create a new transaction with the remainder using account Y as source

TODO: add configuration example

## How to use

TODO: explain how to run the development and production versions

## Original script to verify webhook signature

```php
<?php

declare(strict_types=1);

/*
 * webhook-receiver.php
 * Copyright (c) 2021 james@firefly-iii.org
 *
 * This file is part of Firefly III (https://github.com/firefly-iii).
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

/**
 * A simple script to receive webhooks from Firefly III and verify the signature validity.
 */

/**
 * Define the webhook secret here, otherwise the signature validation will fail.
 */
define('WEBHOOK_SECRET', 'abcdef');

/**
 * Start of script. Here be dragons. Should be no reason to edit beyond this line.
 */
$entry                    = '';
$signatureHash            = null;
$timestamp                = null;
$expectedSignatureVersion = '1';

/**
 * Get values from request (body + signature)
 */
$entityBody = '{}';
$signature  = 't=1610738765,v1=de95f8c28fbeab595d5520205a3b7c2a552811573548d4ad6be786c59a69a495';

/**
 * Explode the signature header to get the necessary data.
 *
 * I know this is terrible code but I'm lazy like that. I'm fairly sure there exists a PHP function like explode_key_value_pairs().
 */
$parts = explode(',', $signature);
foreach ($parts as $row) {
    if ('t=' === substr($row, 0, 2)) {
        $timestamp = substr($row, 2);
    }
    if (sprintf('v%s=', $expectedSignatureVersion) === substr($row, 0, 3)) {
        $signatureHash = trim(substr($row, 3));
    }
}
if (null === $timestamp || null === $signatureHash) {
    echo 'Could not extract valid signature from header :(';
    exit;
}

/**
 * Try to recalculate the signature based on the data. Steal this code.
 */
$payload    = sprintf('%s.%s', $timestamp, $entityBody);
$calculated = hash_hmac('sha3-256', $payload, WEBHOOK_SECRET, false);
$valid      = $calculated === $signatureHash;

/**
 * Put some debug data in a long string.
 */
$return = "Calculated: " . $calculated . "\n";

/**
 * Put more debug data in a long string.
 */
$return .= "\n";
$return .= 'Webhook secret            : ' . WEBHOOK_SECRET . "\n";
$return .= 'Full signature string     : ' . $signature . "\n";
$return .= 'Signature hash            : ' . $signatureHash . "\n";
$return .= 'Signature valid?          : ' . var_export($valid, true) . "\n";
$return .= 'Raw body                  : ' . $entityBody . "\n";
if ($valid) {
    $return .= 'Parsed body (on new line) : ' . "\n";
    $return .= print_r(json_decode($entityBody, true, JSON_THROW_ON_ERROR), true) . "\n";
}
$return .= "\n";


echo $return;
```
