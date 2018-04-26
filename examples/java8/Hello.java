package io.dispatchframework.examples;

import java.util.Map;
import java.util.function.BiFunction;

public class Hello implements BiFunction<Map<String, Object>, Map<String, Object>, String> {
    @Override
    public String apply(Map<String, Object> context, Map<String, Object> payload) {
        final Object name = payload.getOrDefault("name", "Someone");
        final Object place = payload.getOrDefault("place", "Somewhere");

        return String.format("Hello, %s from %s", name, place);
    }
}
