package prompt

const ExtractPrompt = `Extract high-quality retrieval keywords from this conversation.                                     
                                                                                                              
  Goal:                                                                                                                      
  These keywords will be used to search and retrieve this conversation later.                                                
                                                                                                                             
  Rules:                                                                                                                     
  - Focus on specific, meaningful phrases (2–5 words)
  - Include tools, technologies, project names, problems, and decisions                                                      
  - Prefer phrases over single words (e.g., "SQLite local database" instead of "SQLite")
  - Make keywords understandable without the original context                                                                
  - Avoid vague or generic words ("issue", "thing", "problem", "system")                                                     
  - Avoid duplicates                                                                                                         
  - Normalize similar terms                                                                                                  
  - Max 8 keywords                                                                                                           
                                                                                                                             
  Return ONLY a raw JSON array of strings.`
