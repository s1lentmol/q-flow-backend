package com.example.demo.dto;

import com.fasterxml.jackson.annotation.JsonProperty;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class JoinQueueRequest {
    @JsonProperty("group_code")
    private String groupCode;

    @JsonProperty("slot_time")
    private String slotTime;
}
