package com.example.demo.controller;

import com.example.demo.dto.ScheduleJobRequest;
import com.example.demo.dto.ScheduleJobResponse;
import com.example.demo.service.JobSchedulerService;
import org.quartz.SchedulerException;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
// контроллер, принимающий запросы
@RestController
@RequestMapping("/api/jobs")
public class JobSchedulerController {

    @Autowired
    private JobSchedulerService jobSchedulerService;

    @PostMapping("/schedule")
    public ResponseEntity<ScheduleJobResponse> scheduleJob(@RequestBody ScheduleJobRequest request) {
        try {
            if (request.getJobName() == null || request.getJobName().isBlank()) {
                return ResponseEntity.badRequest()
                        .body(new ScheduleJobResponse("error", "Job name is required", null, null));
            }

            if (request.getQueueId() == null) {
                return ResponseEntity.badRequest()
                        .body(new ScheduleJobResponse("error", "Queue ID is required", null, null));
            }

            if (request.getGroupCode() == null || request.getGroupCode().isBlank()) {
                return ResponseEntity.badRequest()
                        .body(new ScheduleJobResponse("error", "Group code is required", null, null));
            }

            if (request.getExecuteAt() == null) {
                return ResponseEntity.badRequest()
                        .body(new ScheduleJobResponse("error", "Execute time is required", null, null));
            }

            jobSchedulerService.scheduleJob(request);

            return ResponseEntity.ok(new ScheduleJobResponse(
                    "success",
                    "Job scheduled successfully",
                    request.getJobName(),
                    request.getExecuteAt().toString()
            ));
        } catch (SchedulerException e) {
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR)
                    .body(new ScheduleJobResponse("error", "Failed to schedule job: " + e.getMessage(), null, null));
        }
    }

    @DeleteMapping("/cancel/{jobName}")
    public ResponseEntity<ScheduleJobResponse> cancelJob(@PathVariable String jobName) {
        try {
            jobSchedulerService.cancelJob(jobName);
            return ResponseEntity.ok(new ScheduleJobResponse(
                    "success",
                    "Job cancelled successfully",
                    jobName,
                    null
            ));
        } catch (SchedulerException e) {
            return ResponseEntity.status(HttpStatus.NOT_FOUND)
                    .body(new ScheduleJobResponse("error", e.getMessage(), jobName, null));
        }
    }
}

