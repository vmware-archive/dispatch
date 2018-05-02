package io.dispatchframework.examples;

import java.util.Map;
import java.util.function.BiFunction;

public class HelloWithClasses implements BiFunction<Map<String, Object>, HelloWithClasses.Payload, HelloWithClasses.Result> {
    @Override
    public Result apply(Map<String, Object> context, Payload payload) {
        final String name = payload.getName() == null ? "Someone" : payload.getName();
        final String place = payload.getPlace() == null ? "Somewhere" : payload.getPlace();

        return new Result(String.format("Hello, %s from %s", name, place));
    }

    // Custom Classes in interface declared as public inner classes
    public class Payload {
        private String name;
        private String place;

        public String getName() {
            return name;
        }

        public String getPlace() {
            return place;
        }
    }

    public class Result {
        private String myField;
        
        public Result(String myField) {
            this.myField = myField;
        }
    }
}
