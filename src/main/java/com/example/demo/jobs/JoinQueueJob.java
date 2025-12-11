package com.example.demo.jobs;

import com.example.demo.service.QueueClient;
import lombok.extern.slf4j.Slf4j;
import org.quartz.Job;
import org.quartz.JobExecutionContext;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Component;
// очередь задач


@Slf4j
@Component
public class JoinQueueJob implements Job {

    @Autowired
    private QueueClient queueClient;

    @Override
    public void execute(JobExecutionContext context) {
        log.info("JoinQueueJob started at {}", context.getFireTime());
        queueClient.joinQueue();
        log.info("JoinQueueJob finished at {}", context.getJobRunTime());
    }
}
