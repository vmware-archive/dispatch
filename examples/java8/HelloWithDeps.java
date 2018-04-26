package io.dispatchframework.examples;

import java.util.Map;
import java.util.function.BiFunction;

import org.joda.time.DateTimeZone;

public class HelloWithDeps implements BiFunction<Map<String, Object>, Map<String, Object>, String> {
    @Override
    public String apply(Map<String, Object> context, Map<String, Object> payload) {
        return String.format("Hello, Someone from timezone %s", DateTimeZone.UTC);
    }
}