FROM swaggerapi/swagger-ui:v5.32.11
COPY docs/openapi.yaml /docs/openapi.yaml
ENV SWAGGER_JSON=/docs/openapi.yaml