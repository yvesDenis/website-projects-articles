package com.custom.metrics.hpa.controller;

import com.custom.metrics.hpa.sender.MessageSender;
import com.fasterxml.jackson.core.JsonProcessingException;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.util.StringUtils;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.Objects;

@RestController
@RequiredArgsConstructor
public class MessageController {

    private final MessageSender messageSender;

    @PostMapping("/message/{name}/counter/{count}")
    public ResponseEntity sendMessageToQueue(@PathVariable String name, @PathVariable Integer count) throws JsonProcessingException {
        if(!this.validateParameters(name,count))
            return ResponseEntity.badRequest().build();
        this.messageSender.sendMessage(name,count);
        return ResponseEntity.status(HttpStatus.ACCEPTED).build();
    }


    @GetMapping(value = "/metrics", produces = "text/plain")
    public String getPendingMessagesNumber(){
        int totalMessages = this.messageSender.pendingMessages();
        return "# HELP messages Number of pending messages in the queue\n"
                + "# TYPE messages gauge\n"
                + "messages " + totalMessages;
    }

    private boolean validateParameters(final String name, final Integer count){
        return StringUtils.hasText(name) && Objects.nonNull(count) && count.compareTo(0) > 0;
    }
}
