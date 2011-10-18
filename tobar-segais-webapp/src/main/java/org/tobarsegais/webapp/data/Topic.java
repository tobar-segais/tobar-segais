package org.tobarsegais.webapp.data;

import javax.xml.stream.XMLStreamConstants;
import javax.xml.stream.XMLStreamException;
import javax.xml.stream.XMLStreamReader;
import javax.xml.stream.XMLStreamWriter;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collection;
import java.util.List;

public class Topic extends Entry {

    private static final long serialVersionUID = 1L;

    public Topic(String label, String href, Topic... children) {
        this(label, href, Arrays.asList(children));
    }

    public Topic(String label, String href, Collection<Topic> children) {
        super(label, href, children);
    }

    public static Topic read(XMLStreamReader reader) throws XMLStreamException {
        if (reader.getEventType() != XMLStreamConstants.START_ELEMENT) {
            throw new IllegalStateException("Expecting a start element");
        }
        if (!"topic".equals(reader.getLocalName())) {
            throw new IllegalStateException("Expecting a <topic> element");
        }
        String label = reader.getAttributeValue(null, "label");
        String href = reader.getAttributeValue(null, "href");
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
        return new Topic(label, href, topics);
    }

    public void write(XMLStreamWriter writer) throws XMLStreamException {
        writer.writeStartElement("topic");
        writer.writeAttribute("label", getLabel());
        writer.writeAttribute("href", getHref());
        for (Topic topic : getChildren()) {
            topic.write(writer);
        }
        writer.writeEndElement();
    }

}
