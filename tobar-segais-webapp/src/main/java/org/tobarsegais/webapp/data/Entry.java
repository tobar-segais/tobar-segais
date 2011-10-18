package org.tobarsegais.webapp.data;

import java.io.Serializable;
import java.util.ArrayList;
import java.util.Collection;
import java.util.Collections;
import java.util.List;

public class Entry implements Serializable {

    private static final long serialVersionUID = 1L;

    private final String label;
    private final List<Topic> children;
    private final String href;

    public Entry(String label, String href, Collection<Topic> children) {
        this.label = label;
        this.children = children == null || children.isEmpty()
                ? Collections.<Topic>emptyList()
                : Collections.unmodifiableList(new ArrayList<Topic>(children));
        this.href = href;
    }

    public String getLabel() {
        return label;
    }

    public List<Topic> getChildren() {
        return children == null ? Collections.<Topic>emptyList() : children;
    }

    public String getHref() {
        return href;
    }

    @Override
    public String toString() {
        final StringBuilder sb = new StringBuilder();
        sb.append("Entry");
        sb.append("{label='").append(getLabel()).append('\'');
        sb.append(", href='").append(getHref()).append('\'');
        sb.append(", children=").append(getChildren());
        sb.append('}');
        return sb.toString();
    }

}
