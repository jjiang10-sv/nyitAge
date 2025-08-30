from collections import deque, defaultdict
from typing import List


def list_str(arr):
    # in case ['a', 'b'] vs ['ab']
    return "-".join(arr)
  
class Solution:
    def deleteDuplicateFolder(self, paths: List[List[str]]) -> List[List[str]]:
        tree = {}
        for path in paths:
            curr = tree
            for name in path:
                curr = curr.setdefault(name, {})

        def encode(node):
            if not node:
                return "()"
            parts = []
            for key, sub in node.items():
                parts.append(key + encode(sub))
            sign = "".join(sorted(parts))
            store[sign].append(node)
            return "(" + sign + ")"

        def remove(nodes):
            for item in nodes:
                item.clear()
                item["#"] = True

        store = defaultdict(list)
        encode(tree)
        for group in store.values():
            if len(group) > 1:
                remove(group)

        def collect(node, path):
            for key, sub in list(node.items()):
                if "#" in sub:
                    continue
                new = path + [key]
                result.append(new)
                collect(sub, new)

        result = []
        collect(tree, [])
        return result
    
class Solution1:
    def deleteDuplicateFolder(self, paths: List[List[str]]) -> List[List[str]]:
        if len(paths) in (0,1):
            return paths

        root_folder = FolderTree("", [])
        tree_folder = root_folder
        # create the tree
        for path in paths:
            i = 0
            for folder in path:
                sub_folder_tree = None
                if folder not in tree_folder.sub_folder_tree_names:
                    sub_folder_tree = FolderTree(folder, path[:i+1])
                    tree_folder.add_sub_folder(sub_folder_tree)
                else:
                    sub_folder_tree = tree_folder.get_sub_folder(folder)
                tree_folder = sub_folder_tree
                i += 1
            tree_folder = root_folder
        # set the sub_folder_name_all for all the tree nodes
        root_folder.set_sub_folder_names_all()
        # mark the node with the same sub_folder_name_all to delete
        same_sub_folder_tree_map = {}
        #deleted_sub_folder_names_all = []
        # BSF on the tree
        folder_queue = deque()
        for sub_folder_tree in tree_folder.sub_folder_trees:
            folder_queue.append((sub_folder_tree,tree_folder))
        while len(folder_queue) > 0:
            
            process_folder = folder_queue.popleft()
            the_folder = process_folder[0]
            the_parent_folder = process_folder[1]
            # if the_folder.path == ["a","b","x"]:
            #     print("the_folder.path == ", the_folder.path)
            process_folder_sub_names = list_str(the_folder.sub_folder_names_all)
            #print("process_folder[0].name ", the_folder.name)
            #print("process_folder_sub_names ", process_folder_sub_names)
            if process_folder_sub_names not in same_sub_folder_tree_map:
                #print("process_folder_sub_names not in ", process_folder_sub_names)
                # if process_folder_sub_names in deleted_sub_folder_names_all:
                #     the_parent_folder.del_sub_folder(the_folder.name)
                #     the_folder.to_del = True
                #     continue

                # not process the leaf folder
                if process_folder_sub_names != "":
                    same_sub_folder_tree_map[process_folder_sub_names] = (the_folder, the_parent_folder)
            else:
                #print("process_folder_sub_names  in ", process_folder_sub_names)
                # del process_folder
                the_parent_folder.del_sub_folder(the_folder.name)
                the_folder.to_del = True
                # del dup 
                dup_folder = same_sub_folder_tree_map[process_folder_sub_names]
                dup_folder[1].del_sub_folder(dup_folder[0].name)
                dup_folder[0].to_del = True
                #deleted_sub_folder_names_all.append(process_folder_sub_names)

            if not the_folder.to_del and not the_parent_folder.to_del:
                for sub_folder_tree in the_folder.sub_folder_trees:
                    folder_queue.append((sub_folder_tree,the_folder))

        result = []
        # reuse the folder_queue for it is empty now. BSF to get all the paths in the remaining nodes
        for sub_folder_tree in root_folder.sub_folder_trees:
            folder_queue.append(sub_folder_tree)
        #print(root_folder.sub_folder_tree_names)
        while len(folder_queue) > 0:           
            process_folder = folder_queue.popleft()
            #print(process_folder.name)
            #print(folder_queue.queue)
            result.append(process_folder.path)
            #print("root_folder.sub_folder_trees ", root_folder.name)
            for sub_folder_tree in process_folder.sub_folder_trees:
                # print(" go down a level")
                # print(sub_folder_tree)
                folder_queue.append(sub_folder_tree)
        return result


class FolderTree:
    
    def __init__(self,name,path):
        self.name = name
        self.sub_folder_trees = []
        self.sub_folder_tree_names = []
        # this tells the sub folder names in the current folder and all the sub folders
        self.sub_folder_names_all = []
        self.path = path
        self.to_del = False

    
    def add_sub_folder(self, sub_folder_tree):
        self.sub_folder_trees.append(sub_folder_tree)
        self.sub_folder_tree_names.append(sub_folder_tree.name)

    def get_sub_folder(self, sub_folder_name):
        sub_folder_idx = None
        for i in range(len(self.sub_folder_tree_names)):
            if self.sub_folder_tree_names[i] == sub_folder_name:
                sub_folder_idx = i
                break
        return self.sub_folder_trees[sub_folder_idx]
    
    # def set_sub_folder_names_all(self):        
    #     for sub_folder_tree in self.sub_folder_trees:
    #         sub_folder_tree.set_sub_folder_names_all()
            
    # DFS
    def set_sub_folder_names_all(self): 
        # ['t', 'd'] matches with ['d', 't']
        if len(self.sub_folder_tree_names) > 0:
            tmp = [x for x in self.sub_folder_tree_names]
            tmp.sort()
            self.sub_folder_names_all.extend(tmp)
        
        for sub_folder_tree in self.sub_folder_trees:
            sub_folder_tree.set_sub_folder_names_all()
            
        for sub_folder_tree in self.sub_folder_trees:
            # this mark means goes to the next level; 
            if len(sub_folder_tree.sub_folder_names_all) > 0:
                self.sub_folder_names_all.append(sub_folder_tree.name)
                self.sub_folder_names_all.extend(sub_folder_tree.sub_folder_names_all)
                
    
    def del_sub_folder(self, sub_folder_name):
        #self.sub_folders = [sub_folder for sub_folder in self.sub_folders if sub_]
        to_del_sub_folder_idx = None
        for i in range(len(self.sub_folder_tree_names)):
            if self.sub_folder_tree_names[i] == sub_folder_name:
                to_del_sub_folder_idx = i
                break
        if to_del_sub_folder_idx is not None:
            self.sub_folder_trees = self.sub_folder_trees[:to_del_sub_folder_idx] + self.sub_folder_trees[to_del_sub_folder_idx+1:]
            self.sub_folder_tree_names = self.sub_folder_tree_names[:to_del_sub_folder_idx] + self.sub_folder_tree_names[to_del_sub_folder_idx+1:]
        #print("del_sub_folder ", self.sub_folder_tree_names)   

solution = Solution()

data = [["b"],["b","d"],["b","d","d"],["b","e"],["c"],["c","d"],["c","e"],["c","e","d"]]

print(solution.deleteDuplicateFolder(data))
[["a"],["c"],["w"],["a","b"],["c","b"],["w","y"],["a","b","x"],["a","b","x","y"]]

[["a"],["c"],["a","b"],["c","b"]]


from collections import deque, defaultdict
from typing import List


def list_str(arr):
    # in case ['a', 'b'] vs ['ab']
    return "-".join(arr)
  
class Solution:
    def deleteDuplicateFolder(self, paths: List[List[str]]) -> List[List[str]]:
        tree = {}
        for path in paths:
            curr = tree
            for name in path:
                curr = curr.setdefault(name, {})

        def encode(node):
            if not node:
                return "()"
            parts = []
            for key, sub in node.items():
                parts.append(key + encode(sub))
            sign = "".join(sorted(parts))
            store[sign].append(node)
            return "(" + sign + ")"

        def remove(nodes):
            for item in nodes:
                item.clear()
                item["#"] = True

        store = defaultdict(list)
        encode(tree)
        for group in store.values():
            if len(group) > 1:
                remove(group)

        def collect(node, path):
            for key, sub in list(node.items()):
                if "#" in sub:
                    continue
                new = path + [key]
                result.append(new)
                collect(sub, new)

        result = []
        collect(tree, [])
        return result
    
class Solution1:
    def deleteDuplicateFolder(self, paths: List[List[str]]) -> List[List[str]]:
        if len(paths) in (0,1):
            return paths

        root_folder = FolderTree("", [])
        tree_folder = root_folder
        # create the tree
        for path in paths:
            i = 0
            for folder in path:
                sub_folder_tree = None
                if folder not in tree_folder.sub_folder_tree_names:
                    sub_folder_tree = FolderTree(folder, path[:i+1])
                    tree_folder.add_sub_folder(sub_folder_tree)
                else:
                    sub_folder_tree = tree_folder.get_sub_folder(folder)
                tree_folder = sub_folder_tree
                i += 1
            tree_folder = root_folder
        # set the sub_folder_name_all for all the tree nodes
        root_folder.set_sub_folder_names_all()
        # mark the node with the same sub_folder_name_all to delete
        same_sub_folder_tree_map = {}
        #deleted_sub_folder_names_all = []
        # BSF on the tree
        folder_queue = deque()
        for sub_folder_tree in tree_folder.sub_folder_trees:
            folder_queue.append((sub_folder_tree,tree_folder))
        while len(folder_queue) > 0:
            
            process_folder = folder_queue.popleft()
            the_folder = process_folder[0]
            the_parent_folder = process_folder[1]
            # if the_folder.path == ["a","b","x"]:
            #     print("the_folder.path == ", the_folder.path)
            process_folder_sub_names = list_str(the_folder.sub_folder_names_all)
            #print("process_folder[0].name ", the_folder.name)
            #print("process_folder_sub_names ", process_folder_sub_names)
            if process_folder_sub_names not in same_sub_folder_tree_map:
                #print("process_folder_sub_names not in ", process_folder_sub_names)
                # if process_folder_sub_names in deleted_sub_folder_names_all:
                #     the_parent_folder.del_sub_folder(the_folder.name)
                #     the_folder.to_del = True
                #     continue

                # not process the leaf folder
                if process_folder_sub_names != "":
                    same_sub_folder_tree_map[process_folder_sub_names] = (the_folder, the_parent_folder)
            else:
                #print("process_folder_sub_names  in ", process_folder_sub_names)
                # del process_folder
                the_parent_folder.del_sub_folder(the_folder.name)
                the_folder.to_del = True
                # del dup 
                dup_folder = same_sub_folder_tree_map[process_folder_sub_names]
                dup_folder[1].del_sub_folder(dup_folder[0].name)
                dup_folder[0].to_del = True
                #deleted_sub_folder_names_all.append(process_folder_sub_names)

            if not the_folder.to_del and not the_parent_folder.to_del:
                for sub_folder_tree in the_folder.sub_folder_trees:
                    folder_queue.append((sub_folder_tree,the_folder))

        result = []
        # reuse the folder_queue for it is empty now. BSF to get all the paths in the remaining nodes
        for sub_folder_tree in root_folder.sub_folder_trees:
            folder_queue.append(sub_folder_tree)
        #print(root_folder.sub_folder_tree_names)
        while len(folder_queue) > 0:           
            process_folder = folder_queue.popleft()
            #print(process_folder.name)
            #print(folder_queue.queue)
            result.append(process_folder.path)
            #print("root_folder.sub_folder_trees ", root_folder.name)
            for sub_folder_tree in process_folder.sub_folder_trees:
                # print(" go down a level")
                # print(sub_folder_tree)
                folder_queue.append(sub_folder_tree)
        return result


class FolderTree:
    
    def __init__(self,name,path):
        self.name = name
        self.sub_folder_trees = []
        self.sub_folder_tree_names = []
        # this tells the sub folder names in the current folder and all the sub folders
        self.sub_folder_names_all = []
        self.path = path
        self.to_del = False

    
    def add_sub_folder(self, sub_folder_tree):
        self.sub_folder_trees.append(sub_folder_tree)
        self.sub_folder_tree_names.append(sub_folder_tree.name)

    def get_sub_folder(self, sub_folder_name):
        sub_folder_idx = None
        for i in range(len(self.sub_folder_tree_names)):
            if self.sub_folder_tree_names[i] == sub_folder_name:
                sub_folder_idx = i
                break
        return self.sub_folder_trees[sub_folder_idx]
    
    # def set_sub_folder_names_all(self):        
    #     for sub_folder_tree in self.sub_folder_trees:
    #         sub_folder_tree.set_sub_folder_names_all()
            
    # DFS
    def set_sub_folder_names_all(self): 
        # ['t', 'd'] matches with ['d', 't']
        if len(self.sub_folder_tree_names) > 0:
            tmp = [x for x in self.sub_folder_tree_names]
            tmp.sort()
            self.sub_folder_names_all.extend(tmp)
        
        for sub_folder_tree in self.sub_folder_trees:
            sub_folder_tree.set_sub_folder_names_all()
            
        for sub_folder_tree in self.sub_folder_trees:
            # this mark means goes to the next level; 
            if len(sub_folder_tree.sub_folder_names_all) > 0:
                self.sub_folder_names_all.append(sub_folder_tree.name)
                self.sub_folder_names_all.extend(sub_folder_tree.sub_folder_names_all)
                
    
    def del_sub_folder(self, sub_folder_name):
        #self.sub_folders = [sub_folder for sub_folder in self.sub_folders if sub_]
        to_del_sub_folder_idx = None
        for i in range(len(self.sub_folder_tree_names)):
            if self.sub_folder_tree_names[i] == sub_folder_name:
                to_del_sub_folder_idx = i
                break
        if to_del_sub_folder_idx is not None:
            self.sub_folder_trees = self.sub_folder_trees[:to_del_sub_folder_idx] + self.sub_folder_trees[to_del_sub_folder_idx+1:]
            self.sub_folder_tree_names = self.sub_folder_tree_names[:to_del_sub_folder_idx] + self.sub_folder_tree_names[to_del_sub_folder_idx+1:]
        #print("del_sub_folder ", self.sub_folder_tree_names)   

solution = Solution()

data = [["b"],["b","d"],["b","d","d"],["b","e"],["c"],["c","d"],["c","e"],["c","e","d"]]

print(solution.deleteDuplicateFolder(data))
[["a"],["c"],["w"],["a","b"],["c","b"],["w","y"],["a","b","x"],["a","b","x","y"]]

[["a"],["c"],["a","b"],["c","b"]]
# not scalable
def deleteDuplicateFolderNotScalable(self, paths: List[List[str]]) -> List[List[str]]:
    if len(paths) in (0,1):
        return paths
    paths.sort()
    print(paths)
    leaf_folder_idxes = []
    paths_len = len(paths)
    prefix_folder = paths[0]
    #leaf_folder = prefix_folder
    paths_map = {list_str(prefix_folder):0}
    parent_children_info = {}
            
    for i in range(1,paths_len):
        folder = paths[i]
        folder_origin = folder
        paths_map[list_str(folder)] = i
        parent = list_str(folder[:-1])
        while parent != "":
            #print("debug paths_map ", paths_map)
            parent_idx = paths_map[parent]
            folder_leaf = folder[-1]
            if parent_idx in parent_children_info:
                # parent_len = len(parent)
                # child = 
                parent_children_info[parent_idx].append(folder_leaf)
            else:
                parent_children_info[parent_idx] = [folder_leaf]
            folder = folder[:-1]
            #print("debug folder ", folder)
            parent = list_str(folder[:-1])
        
        if folder_origin[:-1] != prefix_folder:
            
            leaf_folder_idxes.append(i-1)   
        prefix_folder = folder_origin

    leaf_folder_idxes.append(paths_len-1)

    leaf_map = {}
    to_del_idx = []
    print("parent_children_info === ", parent_children_info)
    for leaf_folder_idx in leaf_folder_idxes:
        leaf_folder = paths[leaf_folder_idx]
        leaf = leaf_folder[-1]
        if leaf not in leaf_map:
            leaf_map[leaf] = leaf_folder_idx
        else:
            # check parent for these two leaves
            path_1 = paths[leaf_folder_idx]
            path_2 = paths[leaf_map[leaf]]
            parent_1 = list_str(path_1[:-1])
            parent_2 = list_str(path_2[:-1])
            parent_1_idx = paths_map[parent_1]
            parent_2_idx = paths_map[parent_2]
            
            if parent_children_info[parent_1_idx] == parent_children_info[parent_2_idx]:
                print("===parent_1_idx parent_2_idx======", parent_1, parent_1_idx, parent_2, parent_2_idx)
                to_del_idx_tmp = [leaf_folder_idx, parent_1_idx, leaf_map[leaf], parent_2_idx]
                to_del_idx.extend(to_del_idx_tmp)
                del parent_children_info[parent_1_idx]
                #del parent_children_info[parent_2_idx]
    print("parent_children_info ==after= ", parent_children_info)
    reversed_map = {}
    for k, v in parent_children_info.items():
        v_str = list_str(v)
        if v_str in reversed_map:
            to_del_idx.append(k)
            to_del_idx.append(reversed_map[v_str])
        else:
            reversed_map[v_str] = k

    result = []
    print("to del idx ", to_del_idx)
    for i in range(len(paths)):
        if i not in to_del_idx:
            result.append(paths[i])
    return result

'''
1, get all the leaf nodes index . put them into an array
2, get all the leaf nodes'a parent index
3, get all the parent's child index
4, loop through the lead nodes index;
    if duplicated , then lookup the parent index; if the parent has the same child nodes, then delete the leaf node and the parent
'''

#   folder = paths[i]
#             paths_map[list_str(folder)] = i
#             if i == len(paths)-1:
#                 leaf = folder[-1]
#                 print("end index, the leaf is ", leaf)
#                 if leaf in leaf_folder_idxes_lookup:
#                     same_leaf_idx = leaf_folder_idxes_lookup[leaf]
#                     parent_idx = paths_map[list_str(folder[:-1])]
#                     same_leaf_idx_path = paths[same_leaf_idx]
#                     same_leaf_idx_parent_idx = paths_map[list_str(same_leaf_idx_path[:-1])]
#                     tmp_to_del_idx = [i,parent_idx,same_leaf_idx,same_leaf_idx_parent_idx]
#                     to_del_idx.extend(tmp_to_del_idx)
#                     print("leaf to delete - end index ", leaf_folder, tmp_to_del_idx)
            
#             if folder[:-1] != prefix_folder:
#                 print("not equal compare --- ", folder[:-1], prefix_folder)
#                 leaf_folder = paths[i-1]
#                 leaf = leaf_folder[-1]
#                 if leaf in leaf_folder_idxes_lookup:
#                     same_leaf_idx = leaf_folder_idxes_lookup[leaf]
#                     parent_idx = paths_map[list_str(leaf_folder[:-1])]
#                     same_leaf_idx_path = paths[same_leaf_idx]
#                     same_leaf_idx_parent_idx = paths_map[list_str(same_leaf_idx_path[:-1])]
#                     tmp_to_del_idx = [i-1,parent_idx,same_leaf_idx,same_leaf_idx_parent_idx]
#                     to_del_idx.extend(tmp_to_del_idx)
#                     print("leaf to delete ", leaf_folder, tmp_to_del_idx)
#                 else:
#                     print("leaf add to lookup ", leaf)
#                     leaf_folder_idxes_lookup[leaf] = i-1
#             prefix_folder = folder
        
#         result = []
#         print("to del idx ", to_del_idx)
#         for i in range(len(paths)):
#             if i not in to_del_idx:
#                 result.append(paths[i])
#         return result


def flatten_nested_dict(nested_dict, separator='.', prefix=''):
    """
    Flatten a nested dictionary into a single-level dictionary.
    
    Args:
        nested_dict: The nested dictionary to flatten
        separator: String to use as separator between nested keys (default: '.')
        prefix: Current key prefix for recursion (used internally)
    
    Returns:
        A flattened dictionary with dot-notation keys
        
    Example:
        Input: {'a': {'b': {'c': 1}, 'd': 2}, 'e': 3}
        Output: {'a.b.c': 1, 'a.d': 2, 'e': 3}
    """
    flattened = {}
    
    for key, value in nested_dict.items():
        new_key = f"{prefix}{separator}{key}" if prefix else key
        
        if isinstance(value, dict) and value:  # Check if it's a non-empty dict
            # Recursively flatten nested dictionaries
            flattened.update(flatten_nested_dict(value, separator, new_key))
        else:
            # Add the key-value pair to the flattened dict
            flattened[new_key] = value
    
    return flattened


def flatten_nested_dict_alternative(nested_dict, separator='.'):
    """
    Alternative implementation using iterative approach with stack.
    
    Args:
        nested_dict: The nested dictionary to flatten
        separator: String to use as separator between nested keys
    
    Returns:
        A flattened dictionary with dot-notation keys
    """
    flattened = {}
    stack = [(nested_dict, '')]
    
    while stack:
        current_dict, prefix = stack.pop()
        
        for key, value in current_dict.items():
            new_key = f"{prefix}{separator}{key}" if prefix else key
            
            if isinstance(value, dict) and value:
                # Push nested dictionary onto stack
                stack.append((value, new_key))
            else:
                # Add the key-value pair to the flattened dict
                flattened[new_key] = value
    
    return flattened


# Test the flattening functions
if __name__ == "__main__":
    # Example nested dictionary
    nested = {
        'a': {
            'b': {
                'c': 1,
                'd': 2
            },
            'e': 3
        },
        'f': 4,
        'g': {
            'h': 5
        }
    }
    
    print("Original nested dictionary:")
    print(nested)
    print()
    
    print("Flattened (recursive):")
    flattened_recursive = flatten_nested_dict(nested)
    print(flattened_recursive)
    print()
    
    print("Flattened (iterative):")
    flattened_iterative = flatten_nested_dict_alternative(nested)
    print(flattened_iterative)
    print()
    
    # Test with custom separator
    print("Flattened with '_' separator:")
    flattened_custom = flatten_nested_dict(nested, separator='_')
    print(flattened_custom)
    
# not scalable
def deleteDuplicateFolderNotScalable(self, paths: List[List[str]]) -> List[List[str]]:
    if len(paths) in (0,1):
        return paths
    paths.sort()
    print(paths)
    leaf_folder_idxes = []
    paths_len = len(paths)
    prefix_folder = paths[0]
    #leaf_folder = prefix_folder
    paths_map = {list_str(prefix_folder):0}
    parent_children_info = {}
            
    for i in range(1,paths_len):
        folder = paths[i]
        folder_origin = folder
        paths_map[list_str(folder)] = i
        parent = list_str(folder[:-1])
        while parent != "":
            #print("debug paths_map ", paths_map)
            parent_idx = paths_map[parent]
            folder_leaf = folder[-1]
            if parent_idx in parent_children_info:
                # parent_len = len(parent)
                # child = 
                parent_children_info[parent_idx].append(folder_leaf)
            else:
                parent_children_info[parent_idx] = [folder_leaf]
            folder = folder[:-1]
            #print("debug folder ", folder)
            parent = list_str(folder[:-1])
        
        if folder_origin[:-1] != prefix_folder:
            
            leaf_folder_idxes.append(i-1)   
        prefix_folder = folder_origin

    leaf_folder_idxes.append(paths_len-1)

    leaf_map = {}
    to_del_idx = []
    print("parent_children_info === ", parent_children_info)
    for leaf_folder_idx in leaf_folder_idxes:
        leaf_folder = paths[leaf_folder_idx]
        leaf = leaf_folder[-1]
        if leaf not in leaf_map:
            leaf_map[leaf] = leaf_folder_idx
        else:
            # check parent for these two leaves
            path_1 = paths[leaf_folder_idx]
            path_2 = paths[leaf_map[leaf]]
            parent_1 = list_str(path_1[:-1])
            parent_2 = list_str(path_2[:-1])
            parent_1_idx = paths_map[parent_1]
            parent_2_idx = paths_map[parent_2]
            
            if parent_children_info[parent_1_idx] == parent_children_info[parent_2_idx]:
                print("===parent_1_idx parent_2_idx======", parent_1, parent_1_idx, parent_2, parent_2_idx)
                to_del_idx_tmp = [leaf_folder_idx, parent_1_idx, leaf_map[leaf], parent_2_idx]
                to_del_idx.extend(to_del_idx_tmp)
                del parent_children_info[parent_1_idx]
                #del parent_children_info[parent_2_idx]
    print("parent_children_info ==after= ", parent_children_info)
    reversed_map = {}
    for k, v in parent_children_info.items():
        v_str = list_str(v)
        if v_str in reversed_map:
            to_del_idx.append(k)
            to_del_idx.append(reversed_map[v_str])
        else:
            reversed_map[v_str] = k

    result = []
    print("to del idx ", to_del_idx)
    for i in range(len(paths)):
        if i not in to_del_idx:
            result.append(paths[i])
    return result

'''
1, get all the leaf nodes index . put them into an array
2, get all the leaf nodes'a parent index
3, get all the parent's child index
4, loop through the lead nodes index;
    if duplicated , then lookup the parent index; if the parent has the same child nodes, then delete the leaf node and the parent
'''

#   folder = paths[i]
#             paths_map[list_str(folder)] = i
#             if i == len(paths)-1:
#                 leaf = folder[-1]
#                 print("end index, the leaf is ", leaf)
#                 if leaf in leaf_folder_idxes_lookup:
#                     same_leaf_idx = leaf_folder_idxes_lookup[leaf]
#                     parent_idx = paths_map[list_str(folder[:-1])]
#                     same_leaf_idx_path = paths[same_leaf_idx]
#                     same_leaf_idx_parent_idx = paths_map[list_str(same_leaf_idx_path[:-1])]
#                     tmp_to_del_idx = [i,parent_idx,same_leaf_idx,same_leaf_idx_parent_idx]
#                     to_del_idx.extend(tmp_to_del_idx)
#                     print("leaf to delete - end index ", leaf_folder, tmp_to_del_idx)
            
#             if folder[:-1] != prefix_folder:
#                 print("not equal compare --- ", folder[:-1], prefix_folder)
#                 leaf_folder = paths[i-1]
#                 leaf = leaf_folder[-1]
#                 if leaf in leaf_folder_idxes_lookup:
#                     same_leaf_idx = leaf_folder_idxes_lookup[leaf]
#                     parent_idx = paths_map[list_str(leaf_folder[:-1])]
#                     same_leaf_idx_path = paths[same_leaf_idx]
#                     same_leaf_idx_parent_idx = paths_map[list_str(same_leaf_idx_path[:-1])]
#                     tmp_to_del_idx = [i-1,parent_idx,same_leaf_idx,same_leaf_idx_parent_idx]
#                     to_del_idx.extend(tmp_to_del_idx)
#                     print("leaf to delete ", leaf_folder, tmp_to_del_idx)
#                 else:
#                     print("leaf add to lookup ", leaf)
#                     leaf_folder_idxes_lookup[leaf] = i-1
#             prefix_folder = folder
        
#         result = []
#         print("to del idx ", to_del_idx)
#         for i in range(len(paths)):
#             if i not in to_del_idx:
#                 result.append(paths[i])
#         return result


def flatten_nested_dict(nested_dict, separator='.', prefix=''):
    """
    Flatten a nested dictionary into a single-level dictionary.
    
    Args:
        nested_dict: The nested dictionary to flatten
        separator: String to use as separator between nested keys (default: '.')
        prefix: Current key prefix for recursion (used internally)
    
    Returns:
        A flattened dictionary with dot-notation keys
        
    Example:
        Input: {'a': {'b': {'c': 1}, 'd': 2}, 'e': 3}
        Output: {'a.b.c': 1, 'a.d': 2, 'e': 3}
    """
    flattened = {}
    
    for key, value in nested_dict.items():
        new_key = f"{prefix}{separator}{key}" if prefix else key
        
        if isinstance(value, dict) and value:  # Check if it's a non-empty dict
            # Recursively flatten nested dictionaries
            flattened.update(flatten_nested_dict(value, separator, new_key))
        else:
            # Add the key-value pair to the flattened dict
            flattened[new_key] = value
    
    return flattened


def flatten_nested_dict_alternative(nested_dict, separator='.'):
    """
    Alternative implementation using iterative approach with stack.
    
    Args:
        nested_dict: The nested dictionary to flatten
        separator: String to use as separator between nested keys
    
    Returns:
        A flattened dictionary with dot-notation keys
    """
    flattened = {}
    stack = [(nested_dict, '')]
    
    while stack:
        current_dict, prefix = stack.pop()
        
        for key, value in current_dict.items():
            new_key = f"{prefix}{separator}{key}" if prefix else key
            
            if isinstance(value, dict) and value:
                # Push nested dictionary onto stack
                stack.append((value, new_key))
            else:
                # Add the key-value pair to the flattened dict
                flattened[new_key] = value
    
    return flattened


# Test the flattening functions
if __name__ == "__main__":
    # Example nested dictionary
    nested = {
        'a': {
            'b': {
                'c': 1,
                'd': 2
            },
            'e': 3
        },
        'f': 4,
        'g': {
            'h': 5
        }
    }
    
    print("Original nested dictionary:")
    print(nested)
    print()
    
    print("Flattened (recursive):")
    flattened_recursive = flatten_nested_dict(nested)
    print(flattened_recursive)
    print()
    
    print("Flattened (iterative):")
    flattened_iterative = flatten_nested_dict_alternative(nested)
    print(flattened_iterative)
    print()
    
    # Test with custom separator
    print("Flattened with '_' separator:")
    flattened_custom = flatten_nested_dict(nested, separator='_')
    print(flattened_custom)