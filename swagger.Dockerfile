FROM swaggerapi/swagger-ui:v5.32.5
COPY docs/openapi.yaml /docs/openapi.yaml
ENV SWAGGER_JSON=/docs/openapi.yaml