package org.tobarsegais.webapp;

import org.tobarsegais.webapp.data.Toc;

import javax.servlet.ServletContext;
import javax.servlet.ServletContextEvent;
import javax.servlet.ServletContextListener;
import javax.xml.stream.XMLStreamException;
import java.net.MalformedURLException;
import java.net.URL;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Set;

/**
 * Created by IntelliJ IDEA.
 * User: stephenc
 * Date: 18/10/2011
 * Time: 09:28
 * To change this template use File | Settings | File Templates.
 */
public class ServletContextListenerImpl implements ServletContextListener{
    public void contextInitialized(ServletContextEvent sce) {
        ServletContext application = sce.getServletContext();
        Map<String,Toc> contents = new LinkedHashMap<String, Toc>();
        for (String path : (Set<String>) application.getResourcePaths("/WEB-INF/bundles")) {
            if (path.endsWith(".jar")) {
                try {
                    URL url = new URL("jar:file://" + application.getRealPath(path) + "!/toc.xml");
                    contents.put(path.substring("/WEB-INF/bundles/".length(), path.lastIndexOf(".jar")), Toc.read(url));
                } catch (XMLStreamException e) {
                    e.printStackTrace();  //To change body of catch statement use File | Settings | File Templates.
                } catch (MalformedURLException e) {
                    e.printStackTrace();  //To change body of catch statement use File | Settings | File Templates.
                }
            }
        }
        application.setAttribute("toc", Collections.unmodifiableMap(contents));
    }

    public void contextDestroyed(ServletContextEvent sce) {
        //To change body of implemented methods use File | Settings | File Templates.
    }
}
