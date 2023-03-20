FROM amazoncorretto:8-alpine-jdk@sha256:7cb2b1e5640cbec45efa8e8d1ee5f90d9bcdb4024728f10181a9dae77d2dff4e as build
WORKDIR /workspace/app
COPY target target
RUN mkdir target/extracted \
    && java -Djarmode=layertools -jar target/*.jar extract --destination target/extracted

FROM amazoncorretto:8-alpine-jdk@sha256:7cb2b1e5640cbec45efa8e8d1ee5f90d9bcdb4024728f10181a9dae77d2dff4e
ARG EXTRACTED=/workspace/app/target/extracted
COPY --from=build ${EXTRACTED}/dependencies/ ./
COPY --from=build ${EXTRACTED}/spring-boot-loader/ ./
COPY --from=build ${EXTRACTED}/snapshot-dependencies/ ./
COPY --from=build ${EXTRACTED}/application/ ./
ENTRYPOINT ["java","org.springframework.boot.loader.JarLauncher"]