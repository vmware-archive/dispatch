///////////////////////////////////////////////////////////////////////
// Copyright (c) 2018 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0
///////////////////////////////////////////////////////////////////////
package io.dispatchframework.examples;

import java.util.Map;
import java.util.function.BiFunction;
import org.springframework.beans.factory.annotation.Qualifier;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

/**
 * A simple example of using Dispatch with a Java function backed by a
 * Spring ApplicationContext.
 */
@Configuration
public class HelloSpring {


    @Bean(name = "noone")
    Person noone() {
        return new Person("Noone", "Nowhere");
    }

    @Bean
    HelloSpringFunction function(@Qualifier("noone") Person noone) {
        return new HelloSpringFunction(noone);
    }

    public class HelloSpringFunction implements BiFunction<Map<Object, Object>, Person, Result> {
        private Person defaultPerson;

        HelloSpringFunction(Person defaultPerson) {
            this.defaultPerson = defaultPerson;
        }

        @Override
        public Result apply(Map<Object, Object> context, Person person) {
        	final String name = person.getName() == null ? defaultPerson.getName() : person.getName();
        	final String place = person.getPlace() == null ? defaultPerson.getPlace() : person.getPlace();
            return new Result("Hello, " + name + " from " + place);
        }
    }

    private class Person {
    	private String name;
    	private String place;

    	public Person(String name, String place) {
    		this.name = name;
    		this.place = place;
    	}

    	public String getName() {
    		return this.name;
    	}

    	public String getPlace() {
    		return this.place;
    	}
    }

    private class Result {
    	private String myField;

    	public Result(String myField) {
    		this.myField = myField;
    	}

    	public String getMyField() {
    		return this.myField;
    	}
    }
}
