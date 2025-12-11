package com.example.demo;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;

@SpringBootApplication
public class QueueSchedulerApplication {
    // запуск бека
    public static void main(String[] args) {
        SpringApplication.run(QueueSchedulerApplication.class, args);
    }
}
