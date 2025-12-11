package com.example.demo.jobs;

import com.example.demo.service.QueueClient;
import org.quartz.Job;
import org.quartz.JobDataMap;
import org.quartz.JobExecutionContext;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class DynamicQueueJob implements Job {

    private static final Logger log = LoggerFactory.getLogger(DynamicQueueJob.class);

    @Override
    public void execute(JobExecutionContext context) {
        JobDataMap dataMap = context.getJobDetail().getJobDataMap();

        String backendBaseUrl = dataMap.getString("backendBaseUrl");
        Long queueId = dataMap.getLong("queueId");
        String groupCode = dataMap.getString("groupCode");
        String slotTime = dataMap.getString("slotTime");

        log.info("DynamicQueueJob started - Job: {}, QueueId: {}, GroupCode: {}",
                context.getJobDetail().getKey().getName(), queueId, groupCode);

        try {
            QueueClient queueClient = new QueueClient(backendBaseUrl);
            queueClient.joinQueue(queueId, groupCode, slotTime);
            log.info("Successfully joined queue {}", queueId);
        } catch (Exception e) {
            log.error("Failed to join queue {}", queueId, e);
        }
    }
}

