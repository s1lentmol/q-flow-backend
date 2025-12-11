package com.example.demo.service;

import com.example.demo.dto.ScheduleJobRequest;
import com.example.demo.jobs.DynamicQueueJob;
import org.quartz.*;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.time.ZoneId;
import java.util.Date;

@Service
public class JobSchedulerService {

    private static final Logger log = LoggerFactory.getLogger(JobSchedulerService.class);

    @Autowired
    private Scheduler scheduler;

    @Value("${queue.backendBaseUrl}")
    private String backendBaseUrl;

    public void scheduleJob(ScheduleJobRequest request) throws SchedulerException {
        String jobName = request.getJobName();
        String jobGroup = "queue-jobs";

        JobDetail jobDetail = JobBuilder.newJob(DynamicQueueJob.class)
                .withIdentity(jobName, jobGroup)
                .usingJobData("backendBaseUrl", backendBaseUrl)
                .usingJobData("queueId", request.getQueueId())
                .usingJobData("groupCode", request.getGroupCode())
                .usingJobData("slotTime", request.getSlotTime() != null ? request.getSlotTime() : "")
                .build();

        Date executeDate = Date.from(request.getExecuteAt().atZone(ZoneId.systemDefault()).toInstant());

        Trigger trigger = TriggerBuilder.newTrigger()
                .withIdentity(jobName + "-trigger", jobGroup)
                .startAt(executeDate)
                .build();

        scheduler.scheduleJob(jobDetail, trigger);

        log.info("Job scheduled: {} to execute at {}", jobName, request.getExecuteAt());
    }

    public void cancelJob(String jobName) throws SchedulerException {
        String jobGroup = "queue-jobs";
        JobKey jobKey = new JobKey(jobName, jobGroup);

        if (scheduler.checkExists(jobKey)) {
            scheduler.deleteJob(jobKey);
            log.info("Job cancelled: {}", jobName);
        } else {
            throw new SchedulerException("Job not found: " + jobName);
        }
    }
}

