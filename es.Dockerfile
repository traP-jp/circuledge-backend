FROM docker.elastic.co/elasticsearch/elasticsearch-wolfi:9.4.4

RUN elasticsearch-plugin install analysis-kuromoji && \
    elasticsearch-plugin install analysis-icu

USER elasticsearch