FROM docker.elastic.co/elasticsearch/elasticsearch-wolfi:9.2.1

RUN elasticsearch-plugin install analysis-kuromoji && \
    elasticsearch-plugin install analysis-icu

USER elasticsearch