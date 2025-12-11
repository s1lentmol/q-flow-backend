package com.example.demo.service;

import com.example.demo.dto.JoinQueueRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.http.HttpEntity;
import org.springframework.http.HttpHeaders;
import org.springframework.http.MediaType;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;
// клиент для добавления в очередь


@Service
public class QueueClient {

    private static final Logger log = LoggerFactory.getLogger(QueueClient.class);

    private final RestTemplate restTemplate;
    private String backendBaseUrl;

    @Value("${queue.queueId:0}")
    private Long queueId;

    @Value("${queue.groupCode:}")
    private String groupCode;

    @Value("${queue.slotTime:}")
    private String slotTime;

    public QueueClient(@Value("${queue.backendBaseUrl}") String backendBaseUrl) {
        this.restTemplate = new RestTemplate();
        this.backendBaseUrl = backendBaseUrl;
    }

    public void joinQueue() {
        joinQueue(queueId, groupCode, slotTime);
    }

    public void joinQueue(Long queueId, String groupCode, String slotTime) {
        String url = backendBaseUrl + "/queues/" + queueId + "/join";

        JoinQueueRequest body = new JoinQueueRequest();
        body.setGroupCode(groupCode);

        if (slotTime != null && !slotTime.isBlank()) {
            body.setSlotTime(slotTime);
        }

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);

        HttpEntity<JoinQueueRequest> entity = new HttpEntity<>(body, headers);

        try {
            ResponseEntity<String> response =
                    restTemplate.postForEntity(url, entity, String.class);

            log.info("Join queue response: status={}, body={}",
                    response.getStatusCode(), response.getBody());
        } catch (Exception e) {
            log.error("Failed to join queue {}", queueId, e);
        }
    }
}
