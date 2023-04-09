package com.custom.metrics.hpa.sender;

import com.custom.metrics.hpa.model.Message;
import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.jms.core.JmsTemplate;
import org.springframework.stereotype.Service;

import java.util.Collections;
import java.util.stream.IntStream;

@Slf4j
@Service
@RequiredArgsConstructor
public class MessageSender {

    private final JmsTemplate jmsTemplate;
    private final ObjectMapper objectMapper;

    @Value("${inbound.queue}")
    private String queueName;

    public void sendMessage(final String messageValue, Integer count) throws JsonProcessingException {
        Message message = new Message(messageValue);
        String jsonValue = this.objectMapper.writeValueAsString(message);
        log.info("Sending message {} times {} to queue - {}" ,count ,jsonValue , this.queueName);
        IntStream.range(0,count).forEach(value -> this.jmsTemplate.convertAndSend(this.queueName,jsonValue));
    }

    public int pendingMessages() {
        return jmsTemplate.browse(this.queueName, (s, qb) -> Collections.list(qb.getEnumeration()).size());
    }
}
