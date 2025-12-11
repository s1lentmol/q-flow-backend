package com.example.demo.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;
// класс с ответом
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ScheduleJobResponse {
    private String status;
    private String message;
    private String jobName;
    private String executeAt;
}

