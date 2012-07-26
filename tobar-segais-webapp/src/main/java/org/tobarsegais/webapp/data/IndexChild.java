package org.tobarsegais.webapp.data;

import javax.xml.stream.XMLStreamException;
import javax.xml.stream.XMLStreamWriter;
import java.io.Serializable;

/**
 * @author stephenc
 * @since 26/07/2012 14:53
 */
public interface IndexChild extends Serializable {

    void write(XMLStreamWriter writer) throws XMLStreamException;
}
