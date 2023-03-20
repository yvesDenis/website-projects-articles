package com.custom.metrics.hpa.receiver;

import com.google.gson.Gson;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.jms.annotation.JmsListener;
import org.springframework.stereotype.Service;

import javax.jms.JMSException;
import javax.jms.Message;
import javax.jms.TextMessage;
import java.util.Map;

@Slf4j
@Service
@RequiredArgsConstructor
public class MessageReceiver {

    @JmsListener(destination = "${inbound.queue}")
    public void receiveMessage(final Message jsonMessage) throws JMSException {
        log.info("Received full message: {}",jsonMessage);
        String response = null;
        if(jsonMessage instanceof TextMessage) {
            String messageData = ((TextMessage)jsonMessage).getText();
            Map map = new Gson().fromJson(messageData, Map.class);
            response  = "Hello ".concat((String) map.get("name"));
        }
        log.info("Received text message: {}", response);
    }
}
