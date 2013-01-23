package org.tobarsegais.webapp;

import javax.servlet.ServletContext;
import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

import org.tobarsegais.webapp.data.Toc;
import org.tobarsegais.webapp.data.TocEntry;

import java.io.IOException;
import java.util.Map;
import java.util.Map.Entry;

/**
 * @author stephenc
 * @since 25/07/2012 11:34
 */
public class DocsServlet extends HttpServlet {

    /**
     * Some bundles can use links that are relative to {@code PLUGINS_ROOT} while others use relative links.
     */
    public static final String PLUGINS_ROOT = "/PLUGINS_ROOT/";

    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse resp) throws ServletException, IOException {
        boolean raw = req.getParameter("raw") != null;
        String topicKey = req.getParameter("topic");
        boolean isTopic = topicKey != null && topicKey.length() > 0;
        
        String path = req.getPathInfo();
        if (path == null) {
            path = req.getServletPath();
        }
        int index = path.indexOf(PLUGINS_ROOT);
        if (index != -1) {
            path = path.substring(index + PLUGINS_ROOT.length() - 1);
        }

        if( isTopic ){
        	path = findTopicPath(topicKey);
        }
        
        if (path.equals("/docs")) {
            resp.sendRedirect("/docs/");
            return;
        }

        int endOfFileName = path.indexOf('#');
        endOfFileName = endOfFileName == -1 ? path.length() : endOfFileName;
        int startOfFileName = path.lastIndexOf('/', endOfFileName);
        startOfFileName = startOfFileName == -1 ? 0 : startOfFileName + 1;
        String fileName = path.substring(startOfFileName, endOfFileName);

        if (raw || (!fileName.toLowerCase().endsWith(".htm") && !fileName.toLowerCase().endsWith(".html") && !fileName
                .isEmpty())) {
            req.getRequestDispatcher("/content" + path).forward(req, resp);
        } else {
        	if( isTopic ){
        		req.setAttribute("content", path );
        	} else {
        		req.setAttribute("content", req.getPathInfo());
        	}
            req.getRequestDispatcher("/WEB-INF/docs-template.jsp").forward(req, resp);
        }
    }
    
    /**
     * search full path for given topic href
     * 
     * if topic key place more than one bundle, returns first found
     * 
     * @param topicKey
     * @return
     */
    protected String findTopicPath( String topicKey ){
    	ServletContext application = getServletContext();
    	Map<String, Toc> contents = ServletContextListenerImpl.getTablesOfContents(application);
    	for( Entry<String, Toc> entry :  contents.entrySet() ){
    		TocEntry tocEntry = entry.getValue().lookupTopic(topicKey);
    		if( tocEntry != null ){
    			return "/" + entry.getKey() + "/" + tocEntry.getHref();
    		}
    	}
    	return "/docs";
    }
}
