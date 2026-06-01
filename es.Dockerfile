FROM docker.elastic.co/elasticsearch/elasticsearch-wolfi:9.4.2

RUN elasticsearch-plugin install analysis-kuromoji && \
    elasticsearch-plugin install analysis-icu

USER elasticsearch