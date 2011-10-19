package org.tobarsegais.webapp.data;

import com.thoughtworks.xstream.XStream;
import org.apache.commons.io.IOUtils;

import javax.xml.stream.XMLInputFactory;
import javax.xml.stream.XMLStreamConstants;
import javax.xml.stream.XMLStreamException;
import javax.xml.stream.XMLStreamReader;
import javax.xml.stream.XMLStreamWriter;
import java.io.IOException;
import java.io.InputStream;
import java.net.URL;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.List;

public class Toc extends Entry {

    private static final long serialVersionUID = 1L;

    public Toc(String label, String topic, Topic... children) {
        this(label, topic, Arrays.asList(children));
    }

    public Toc(String label, String topic, Collection<Topic> children) {
        super(label, topic, children);
    }

    public static Toc read(InputStream inputStream) throws XMLStreamException {
        try {
            return read(XMLInputFactory.newInstance().createXMLStreamReader(inputStream));
        } finally {
            IOUtils.closeQuietly(inputStream);
        }
    }

    public static Toc read(XMLStreamReader reader) throws XMLStreamException {
        while (reader.hasNext() && !reader.isStartElement()) {
            reader.next();
        }
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"toc".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <toc> element");
        }
        String label = reader.getAttributeValue(null, "label");
        String topic = reader.getAttributeValue(null, "topic");
        List<Topic> topics = new ArrayList<Topic>();
        outer:
        while (reader.hasNext()) {
            switch (reader.next()) {
                case XMLStreamConstants.START_ELEMENT:
                    topics.add(Topic.read(reader));
                    break;
                case XMLStreamConstants.END_ELEMENT:
                    break outer;
            }
        }
        return new Toc(label, topic, topics);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("toc");
        writer.writeAttribute("label", getLabel());
        writer.writeAttribute("topic", getHref());
        for (Topic topic: getChildren()) {
           topic.write(writer);
        }
        writer.writeEndElement();
    }

}
