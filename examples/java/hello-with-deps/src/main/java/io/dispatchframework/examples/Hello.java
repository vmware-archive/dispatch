/*
 * Copyright (c) 2018 VMware, Inc. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package io.dispatchframework.examples;

import java.util.Map;
import java.util.function.BiFunction;

public class Hello implements BiFunction<Map<String, Object>, Map<String, Object>, String> {
    public String apply(Map<String, Object> context, Map<String, Object> payload) {
        final Object name = payload.getOrDefault("name", "Someone");
        final Object place = payload.getOrDefault("place", "Somewhere");

        return String.format("Hello, %s from %s", name, place);
    }
}
